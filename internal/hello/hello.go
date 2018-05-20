package hello

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
    "strings"
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
    Tcp_tcp_len string
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
    simplesrc := simpleHost(p.Layers.Ip.Ip_ip_src_host)
    simpledst := simpleHost(p.Layers.Ip.Ip_ip_dst_host)
    return simplesrc, simpledst
}

func (p Packet) size() int {
    // TODO detect size of packet... radio layer?
    size, err := strconv.Atoi(p.Layers.Tcp.Tcp_tcp_len)
    if err != nil {
        return 0
    }
    return size
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
    Timestamp int64
    Count int
}

type ConversationHist struct {
    Source string
    Destination string
    Traffic []TrafChunk
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

func isIPv4(h string) bool {
    hostparts := strings.Split(h, ".")
    _, err := strconv.Atoi(hostparts[len(hostparts) - 1])
    return err == nil
}

/**
 *  Hackish -- truncate subdomains
 */
func simpleHost(h string) string {
    if isIPv4(h) {
        return h
    } else {
        hostparts := strings.Split(h, ".")
        start := 0
        if len(hostparts) > 2 {
            start = len(hostparts) - 2
        }
        return strings.Join(hostparts[start:], ".")
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

func pkts2hist(pstream <- chan Packet, hstream chan <- []ConversationHist, delta int64, windows int) {
    go func() {
        // ip_addr: position in dh
        devices := make(map[string]int)
        dh := make([]ConversationHist, 0)
        var offset int64 = -1
        var lastsent int64 = -1
        pkts := 0

        for {
            select {
            case packet := <- pstream:
                pkts += 1
                fmt.Println("processing packet", pkts)
                tm, err := strconv.ParseInt(packet.Timestamp, 10, 64)
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
                        hstream <- dh
                        for index, convo := range dh {
                            chunk := TrafChunk{offset, 0}
                            dh[index].Traffic = append(convo.Traffic, chunk)
                            traflen := len(dh[index].Traffic)
                            if traflen > windows {
                                dh[index].Traffic = dh[index].Traffic[traflen-windows:]
                            }
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
                if offset > lastsent {
                    fmt.Println("exporting histogram")
                    hstream <- dh
                    lastsent = offset
                }
            }
        }
    }()
}

func Main() {
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

    var resms int64 = 1000
    hist_windows := 250

    pstream := make(chan Packet, 1000)
    hstream := make(chan []ConversationHist)

    si2pkts(pstream)
    pkts2hist(pstream, hstream, resms, hist_windows)

    // could be part of pkts2hist, and just write to disk
    for {
        var lastts int64 = 0
        select {
        case hist := <- hstream:
            var newts int64 = 0
            if len(hist) > 0 {
                newts = hist[0].Traffic[len(hist[0].Traffic) - 1].Timestamp
            }
            if newts != lastts {
                savehist(hist, os.Getenv("GOPATH") + "/src/pi-eye/web/visualization/hist.json")
                lastts = newts
            }
        }
    }
}