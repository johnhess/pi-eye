package tshark

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
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
func (p Packet) Fromto() (string, string) {
    simplesrc := simpleHost(p.Layers.Ip.Ip_ip_src_host)
    simpledst := simpleHost(p.Layers.Ip.Ip_ip_dst_host)
    return simplesrc, simpledst
}

func (p Packet) Size() int {
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
func Si2pkts(c chan <- Packet) {
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

