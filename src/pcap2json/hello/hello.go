package main

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
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

func str2pkt(s string) (Packet, error) {
    pkt := Packet{}
    if err := json.Unmarshal([]byte(s), &pkt); err != nil {
        return Packet{}, errors.New("malformed JSON")
    }
    return pkt, nil;
}

type TrafChunk struct {
    Timestamp int
    Count int
}

func traffichist(pkts []Packet, delta int) []TrafChunk {
    // Create and fill chunks as long as there are still packets in the array.
    pkti := 0
    var chunks []TrafChunk
    // First chunk starts at the time of the first packet
    offset, err := strconv.Atoi(pkts[0].Timestamp)
    if err != nil {
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

func read() {
    fmt.Println("Attempting to parse EK data from file.")
    file, err := os.Open("/Users/johnhess/Dropbox/hackamajig/netcaptures/stream.ek")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    var pkts []Packet

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        pkt, err := str2pkt(line)
        // if we hit a malformed row, we just skip it later.
        if err != nil {
            fmt.Println("caught end of file or malformed json.")
        } else {
            if (Ip{}) != pkt.Layers.Ip || (Dns{}) != pkt.Layers.Dns {
                pkts = append(pkts, pkt)
            }
        }
    }

    if err := scanner.Err(); err != nil {
        panic(err)
    }

    hist := traffichist(pkts, 1000)
    var histstart int
    if len(hist) < 1920 {
        histstart = 0
    } else {
        histstart = len(hist) - 1920
    }
    savehist(hist[histstart:], "/Users/johnhess/Dropbox/hackamajig/networkviz/hist.json")
    fmt.Println("Done.")
}

func returntwo() int {
    return 2
}

func main() {
    // just run again if we eventually complete read()
    for {
        read()
    }
}