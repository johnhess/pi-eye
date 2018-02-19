package main

import (
    "bufio"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "runtime/pprof"
    "strconv"
    "time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

type Ip struct {
    Ip_ip_dst_host string
    Ip_ip_dst string
    Ip_ip_src_host string
    Ip_ip_src string
}

func (ip Ip) String() string {
    if ip.Ip_ip_dst_host != "" {
        return fmt.Sprintf(
            "IP Information: DST: %s SRC: %s", 
            ip.Ip_ip_dst_host, 
            ip.Ip_ip_src_host)
    }
    return ""
}

type Dns struct {
    Text_dns_qry_name string
}

func (dns Dns) String() string {
    if dns.Text_dns_qry_name != "" {
        return fmt.Sprintf("DNS Information: %s", dns.Text_dns_qry_name)
    }
    return ""
}

type Tcp struct {
    Tcp_analysis_tcp_analysis_bytes_in_flight string
    Tcp_tcp_dstport string
    Tcp_tcp_srcport string
}

func (tcp Tcp) String() string {
    if tcp != (Tcp{}) {
        return fmt.Sprintf(
            "TCP (%s to %s) Size: %s", 
            tcp.Tcp_tcp_srcport,
            tcp.Tcp_tcp_dstport,
            tcp.Tcp_analysis_tcp_analysis_bytes_in_flight)
    }
    return ""
}

type Layers struct {
    Dns Dns
    Ip Ip
    Tcp Tcp
    // TODO capture wlan, too when at wireshark version 2.5+ (so field is not repeated)
}

type Packet struct {
    Timestamp string
    Layers Layers
}

/*
 * The resolved host name of the requesting party (src/dst).
 *
 * By convention, if communicating over TCP to a port < 49151, the packet src
 * is the requester, otherwise, the dest.
 */
func (p Packet) fromto() (string, string) {
    return p.Layers.Ip.Ip_ip_src_host, p.Layers.Ip.Ip_ip_dst_host
}

func (p Packet) size() int {
    // TODO detect size of packet... radio layer?
    return 1
}

func bytes2packet(b []byte) (Packet, error) {
    pkt := Packet{}
    if err := json.Unmarshal(b, &pkt); err != nil {
        fmt.Println(string(b))
        return Packet{}, errors.New("malformed JSON")
    }
    return pkt, nil;
}

type TrafChunk struct {
    Timestamp int
    Count int
}

type ConversationHist struct {
    Source string
    Destination string
    Traffic []TrafChunk
}

// TODO: stream new results to here, return only the tail window.
func mdhist(pkts []Packet, delta int) []ConversationHist {
    // ip_addr: position in dh
    devices := make(map[string]int)
    dh := make([]ConversationHist, 0)
    
    if len(pkts) == 0 {
        return dh
    }
    pkti := 0
    // First chunk starts at the time of the first packet
    offset, err := strconv.Atoi(pkts[0].Timestamp)
    if err != nil {
        fmt.Println(pkts)
        panic(err)
    }
    for {
        // Initialize Chunks for all conversations
        for index, convo := range dh {
            chunk := TrafChunk{offset, 0}
            dh[index].Traffic = append(convo.Traffic, chunk)
        }
        for {
            // end of pkts
            if pkti >= len(pkts) {break}
            tm, err := strconv.Atoi(pkts[pkti].Timestamp)
            if err != nil {
                panic(err)
            }
            src, dest := pkts[pkti].fromto()
            convo := src + ":" + dest
            if tm < offset + delta {
                if _, ok := devices[convo]; !ok {
                    // device not yet in map or array
                    newdh := ConversationHist{src, dest, []TrafChunk{TrafChunk{offset, 0}}}
                    dh = append(dh, newdh)
                    devices[convo] = len(dh) - 1
                }
                dtraf := dh[devices[convo]].Traffic
                size := pkts[pkti].size()
                dtraf[len(dtraf) - 1].Count += size
                pkti++
            } else {break} // end of chunk
        }
        offset += delta
        // finally stop if we're out of packets
        if pkti >= len(pkts) {break}
    }
    return dh
}

func savehist(hist interface{}, f string) {
    out, err := json.Marshal(hist)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile(f, []byte(string(out)), 0644)
    if err != nil {
        panic(err)
    }
}

/*
 *  Streams interesting packets from stdin to a channel.
 *
 *  Returns immediately.
 */
func si2pkts(c chan <- Packet) {
    go func() {
        stdin := bufio.NewReader(os.Stdin)
        for {
            // Grab lines from the file
            line, err := stdin.ReadString('\n')
            if err != nil {
                switch err {
                case io.EOF:
                    time.Sleep(1 * time.Millisecond)
                default:
                    panic(err)
                }
            } else {
                // Make a packet
                if pkt, err := bytes2packet([]byte(line)); err != nil {
                    panic(err)
                } else if (Ip{}) != pkt.Layers.Ip || (Dns{}) != pkt.Layers.Dns {
                    c <- pkt
                }
            }
        }
    }()
}

func pkts2hist(pstream <- chan Packet, hstream chan <- []ConversationHist) {
    go func() {
        // ip_addr: position in dh
        devices := make(map[string]int)
        dh := make([]ConversationHist, 0)
        offset := -1;
        delta := 1000  // Milliseconds
        pkts := 0

        for {
            select {
            case packet := <- pstream:
                pkts += 1
                fmt.Println("processing packet", pkts)
                tm, err := strconv.Atoi(packet.Timestamp)
                if err != nil {
                    panic(err)
                }
                if offset == -1 {
                    offset = tm
                    for index, convo := range dh {
                        chunk := TrafChunk{offset, 0}
                        dh[index].Traffic = append(convo.Traffic, chunk)
                    }
                } else if tm >= offset + delta {
                    for {
                        offset = offset + delta
                        for index, convo := range dh {
                            chunk := TrafChunk{offset, 0}
                            dh[index].Traffic = append(convo.Traffic, chunk)
                        }
                        if tm <= offset + delta {break}
                    }
                }
                src, dest := packet.fromto()
                convo := src + ":" + dest
                if _, ok := devices[convo]; !ok {
                    // device not yet in map or array
                    newdh := ConversationHist{src, dest, []TrafChunk{TrafChunk{offset, 0}}}
                    dh = append(dh, newdh)
                    devices[convo] = len(dh) - 1
                }
                dtraf := dh[devices[convo]].Traffic
                dtraf[len(dtraf) - 1].Count += packet.size()
            default:
                hstream <- dh
                time.Sleep(10 * time.Millisecond)
            }
        }
    }()
}

func main() {
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }

    pstream := make(chan Packet, 1000)
    hstream := make(chan []ConversationHist)

    si2pkts(pstream)
    pkts2hist(pstream, hstream)

    // could be part of pkts2hist, and just write to disk
    for {
        select {
        case hist := <- hstream:
            savehist(hist, "/Users/johnhess/Dropbox/hackamajig/networkviz/hist.json")
        }
    }
}