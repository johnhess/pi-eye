package main

import (
    "fmt"
    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
)

func read() {
    fmt.Println("Attempting to parse pcap")
    if handle, err := pcap.OpenOffline("/Users/johnhess/test.pcap"); err != nil {
    // if handle, err := pcap.OpenOffline("/Users/johnhess/Downloads/dns.cap"); err != nil {
        panic(err)
    } else {
        packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
        for packet := range packetSource.Packets() {
            fmt.Println(packet)
        }
    }

}

func returntwo() int {
    return 2
}

func main() {
    fmt.Println("Hello, World.")
    read()
}