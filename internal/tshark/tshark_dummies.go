package tshark

import (
)

var Dummy_packets = []Packet{
    // Conversations
    // dest1 --> src1 (2 packets)
    // dest1 --> src2
    // dest2 --> src2
    Packet{"1000000000000", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src1", "src1"},
            Tcp{"100", "443", "5555", "10"}}}, 
    Packet{"1000000004001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src2", "src2"},
            Tcp{"100", "443", "5555", "10"}}}, 
    Packet{"1000000004001", Layers{
            Dns{},
            Ip{"dest1", "dest1", "src1", "src1"},
            Tcp{"100", "443", "5555", "10"}}}, 
    Packet{"1000000009001", Layers{
            Dns{},
            Ip{"dest2", "dest2", "src2", "src2"},
            Tcp{"100", "443", "5555", "10"}}},
}