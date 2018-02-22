package main

import (
    "fmt"
    "testing"
)

var pkts = []Packet{
    // Conversations
    // dest1 --> src1 (2 packets)
    // dest1 --> src2
    // dest2 --> src2
    Packet{"1000000000000", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src1", "src1"},
            Tcp{"100", "443", "5555"}}}, 
    Packet{"1000000004001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src2", "src2"},
            Tcp{"100", "443", "5555"}}}, 
    Packet{"1000000004001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src1", "src1"},
            Tcp{"100", "443", "5555"}}}, 
    Packet{"1000000009001", Layers{
            Dns{},
            Ip{"dest2", "dest2", "src2", "src2"},
            Tcp{"100", "443", "5555"}}},
}


func TestStreamHistGen(t *testing.T) {
    pstream := make(chan Packet, 1)
    hstream := make(chan []ConversationHist)

    pkts2hist(pstream, hstream, 1000, 100)

    pstream <- pkts[0]
    hist := <- hstream
    if len(hist) != 1 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(hist)))
    }
    pstream <- pkts[1]
    hist = <- hstream
    if len(hist) != 2 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(hist)))
    }
    if len(hist[0].Traffic) != 5 {
        fmt.Println(hist)
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(hist[0].Traffic)))
    }
}

func TestHistTruncation(t *testing.T) {
    pstream := make(chan Packet, 1000)
    hstream := make(chan []ConversationHist)

    pkts2hist(pstream, hstream, 1000, 3)

    pstream <- pkts[0]
    pstream <- pkts[1]
    hist := <- hstream
    if len(hist[0].Traffic) != 3 {
        fmt.Println(hist)
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(hist[0].Traffic)))
    }
}

func TestHostLeavesLocalIPsUntouched(t *testing.T) {
    clean := simpleHost("192.168.1.106")
    if clean != "192.168.1.106" {t.Error(clean)}
    clean = simpleHost("172.20.1.1")
    if clean != "172.20.1.1" {t.Error(clean)}
}

func TestHostTruncatesSubdomains(t *testing.T) {
    clean := simpleHost("this.com")
    if clean != "this.com" {t.Error(clean)}
    clean = simpleHost("sub.this.com")
    if clean != "this.com" {t.Error(clean)}
}