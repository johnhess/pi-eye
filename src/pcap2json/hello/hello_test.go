package main

import (
    "fmt"
    "testing"
)

var pkts = []Packet{
    Packet{"1000000000000", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src1", "src1"},
            Tcp{}}}, 
    Packet{"1000000004001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src2", "src2"},
            Tcp{}}}, 
    Packet{"1000000004001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src1", "src1"},
            Tcp{}}}, 
    Packet{"1000000009001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src2", "src2"},
            Tcp{}}},
}

func TestHist(t *testing.T) {
    hist := traffichist(pkts, 1000)
    if len(hist) != 10 {
        t.Errorf(fmt.Sprintf("Wrong number of slices in hist: %d", len(hist)))
    }
    if hist[0].Count != 1 {
        t.Errorf("first chunk should have seen one packet")
    }
    if hist[1].Count != 0 {
        t.Errorf("second chunk should have seen no packets")
    }
    if hist[4] != (TrafChunk{1000000004000, 2}) {
        t.Errorf("didnt catch double packet window")
    }
}

func TestMultiDeviceHist(t *testing.T) {
    mdh := mdhist(pkts, 1000)
    fmt.Println(mdh)
    if len(mdh) != 2 {
        t.Errorf(fmt.Sprintf("Wrong number of devices: %d", len(mdh)))
    }
    if mdh[0].Device != "src1" || mdh[1].Device != "src2" {
        t.Error("Incorrect device names")
    }
    if len(mdh[0].Traffic) != 10 {
        fmt.Println(mdh[0])
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(mdh[0].Traffic)))
    }
    if len(mdh[1].Traffic) != 6 {
        t.Error(fmt.Sprintf("Incorrect traffic length: %d", len(mdh[1].Traffic)))
    }
}

/*

Desired output 

[
    {
        device: "192.168.0.103",
        traffic: []  //traffichist
    }
    {
        device: "192.168.0.123",
        traffic: []  //traffichist
    }
]

*/