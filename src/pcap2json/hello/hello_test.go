package main

import (
    "fmt"
    "testing"
)

var pkts = []Packet{
    Packet{"1000000000000", Layers{}}, 
    Packet{"1000000004001", Layers{}}, 
    Packet{"1000000004001", Layers{}}, 
    Packet{"1000000009001", Layers{}}}

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