package hists

import (
    "fmt"
    "testing"
    "pi-eye/internal/tshark"
)

var pkts = []tshark.Packet{
    // Conversations
    // dest1 --> src1 (2 packets)
    // dest1 --> src2
    // dest2 --> src2
    tshark.Packet{"1000000000000", tshark.Layers{
            tshark.Dns{},
            tshark.Ip{"dest1", "dest1", "src1", "src1"},
            tshark.Tcp{"100", "443", "5555", "10"}}}, 
    tshark.Packet{"1000000004001", tshark.Layers{
            tshark.Dns{},
            tshark.Ip{"dest1", "dest1", "src2", "src2"},
            tshark.Tcp{"100", "443", "5555", "10"}}}, 
    tshark.Packet{"1000000004001", tshark.Layers{
            tshark.Dns{},
            tshark.Ip{"dest1", "dest1", "src1", "src1"},
            tshark.Tcp{"100", "443", "5555", "10"}}}, 
    tshark.Packet{"1000000009001", tshark.Layers{
            tshark.Dns{},
            tshark.Ip{"dest2", "dest2", "src2", "src2"},
            tshark.Tcp{"100", "443", "5555", "10"}}},
}


func TestStreamHistGen(t *testing.T) {
    pstream := make(chan tshark.Packet, 1)
    hstream := make(chan []ConversationHist)

    pkts2hist(pstream, hstream, 1000, 100)

    pstream <- pkts[0]
    hist := <- hstream
    if len(hist) != 1 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(hist)))
    }
    pstream <- pkts[1]
    // Smooth exports mean that each window results in a new hist
    hist = <- hstream
    hist = <- hstream
    hist = <- hstream
    hist = <- hstream
    hist = <- hstream
    if len(hist) != 2 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(hist)))
    }
    if len(hist[0].Traffic) != 5 {
        fmt.Println(hist)
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(hist[0].Traffic)))
    }
    if hist[0].Traffic[0].Count != 10 {
        t.Error(fmt.Sprintf("Incorrect traffic volume: %d", hist[0].Traffic[0].Count))        
    }
}

func TestHistTruncation(t *testing.T) {
    pstream := make(chan tshark.Packet, 1000)
    hstream := make(chan []ConversationHist)

    pkts2hist(pstream, hstream, 1000, 3)

    pstream <- pkts[0]
    pstream <- pkts[1]
    hist := <- hstream
    hist = <- hstream
    hist = <- hstream
    hist = <- hstream
    hist = <- hstream
    if len(hist[0].Traffic) != 3 {
        fmt.Println(hist)
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(hist[0].Traffic)))
    }
}