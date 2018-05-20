package convos

import (
    "fmt"
    "testing"
    "pi-eye/internal/tshark"
)


func TestConvo(t *testing.T) {
    pstream := make(chan tshark.Packet, 1)
    cstream := make(chan map[string]Conversation)

    pkts2convos(pstream, cstream)

    pstream <- tshark.Dummy_packets[0]
    convos := <- cstream
    if len(convos) != 1 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(convos)))
    }
}

func TestConvos(t *testing.T) {
    pstream := make(chan tshark.Packet, 1)
    cstream := make(chan map[string]Conversation)

    pkts2convos(pstream, cstream)

    pstream <- tshark.Dummy_packets[0]
    convos := <- cstream
    pstream <- tshark.Dummy_packets[1]
    convos = <- cstream
    if len(convos) != 2 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(convos)))
    }
}
