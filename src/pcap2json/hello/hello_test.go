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

func TestConvoHist(t *testing.T) {
    mdh := mdhist(pkts, 1000)
    if len(mdh) != 3 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(mdh)))
    }
    if mdh[0].Source != "src1" || mdh[1].Destination != "dest1" {
        t.Error("Incorrect src/dest names")
    }
    if len(mdh[0].Traffic) != 10 {
        fmt.Println(mdh[0])
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(mdh[0].Traffic)))
    }
    if mdh[0].Traffic[0].Count != 1 {
        t.Error(fmt.Sprintf("Incorrect traffic value: %d", mdh[1].Traffic[0].Count))
    }
}

func TestStreamHistGen(t *testing.T) {
    pstream := make(chan Packet, 1000)
    hstream := make(chan []ConversationHist)

    pkts2hist(pstream, hstream)    

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