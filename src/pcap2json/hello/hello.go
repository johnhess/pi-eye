package main

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "strconv"
    "time"
)

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
}

func (tcp Tcp) String() string {
    if tcp != (Tcp{}) {
        return fmt.Sprintf(
            "TCP Size: %s", 
            tcp.Tcp_analysis_tcp_analysis_bytes_in_flight)
    }
    return ""
}

type Layers struct {
    Dns Dns
    Ip Ip
    Tcp Tcp
}

type Packet struct {
    Timestamp string
    Layers Layers
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

type DeviceHist struct {
    Device string
    Traffic []TrafChunk
}

func traffichist(pkts []Packet, delta int) []TrafChunk {
    var chunks []TrafChunk
    if len(pkts) == 0 {
        return chunks
    }
    pkti := 0
    // First chunk starts at the time of the first packet
    offset, err := strconv.Atoi(pkts[0].Timestamp)
    if err != nil {
        fmt.Println(pkts)
        panic(err)
    }
    for {
        chunk := TrafChunk{offset, 0}
        for {
            // end of pkts
            if pkti >= len(pkts) {break}
            tm, err := strconv.Atoi(pkts[pkti].Timestamp)
            if err != nil {
                panic(err)
            }
            if tm < offset + delta {
                chunk.Count++
                pkti++
            } else {break} // end of chunk
        }
        chunks = append(chunks, chunk)
        offset += delta
        // finally stop if we're out of packets
        if pkti >= len(pkts) {break}
    }
    return chunks
}

func mdhist(pkts []Packet, delta int) []DeviceHist {
    // ip_addr: position in dh
    devices := make(map[string]int)
    dh := make([]DeviceHist, 0)
    
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
        for index, device := range dh {
            chunk := TrafChunk{offset, 0}
            dh[index].Traffic = append(device.Traffic, chunk)
        }
        for {
            // end of pkts
            if pkti >= len(pkts) {break}
            tm, err := strconv.Atoi(pkts[pkti].Timestamp)
            if err != nil {
                panic(err)
            }
            device := pkts[pkti].Layers.Ip.Ip_ip_src
            if tm < offset + delta {
                if _, ok := devices[device]; !ok {
                    // device not yet in map or array
                    newdh := DeviceHist{device, []TrafChunk{TrafChunk{offset, 0}}}
                    dh = append(dh, newdh)
                    devices[device] = len(dh) - 1
                }
                dtraf := dh[devices[device]].Traffic
                dtraf[len(dtraf) - 1].Count++
                pkti++
            } else {break} // end of chunk
        }
        offset += delta
        // finally stop if we're out of packets
        if pkti >= len(pkts) {break}
    }
    return dh
}

func savehist(hist []TrafChunk, f string) {
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
                } else {
                    c <- pkt
                }
            }
        }
    }()
}

func main() {
    stream := make(chan Packet, 1000)

    si2pkts(stream)

    var pkts []Packet
    lastExportLen := 0
    for {
        select {
        case pkt := <-stream:
            // Toss uninteresting packets
            if (Ip{}) != pkt.Layers.Ip || (Dns{}) != pkt.Layers.Dns {
                pkts = append(pkts, pkt)
            }
        default:
            // No packets?  Export.  Could probably be in a goroutine.
            if len(pkts) > lastExportLen {
                hist := traffichist(pkts, 100)
                var histstart int
                if len(hist) < 1920 {
                    histstart = 0
                } else {
                    histstart = len(hist) - 1920
                }
                savehist(hist[histstart:], "/Users/johnhess/Dropbox/hackamajig/networkviz/hist.json")
                lastExportLen = len(pkts)
                fmt.Println(fmt.Sprintf("Exported %d packets.", lastExportLen))
            }
            time.Sleep(1 * time.Millisecond)
        }
    }
}