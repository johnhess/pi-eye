package convos

import (
    "fmt"
    "testing"
    "pi-eye/internal/tshark"
)


func TestConvos(t *testing.T) {
    pstream := make(chan tshark.Packet, 1)
    cstream := make(chan map[string]Conversation)

    pkts2convos(pstream, cstream)

    pstream <- tshark.Dummy_packets[0]
    convos := <- cstream
    if len(convos) != 1 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(convos)))
    }
    pstream <- tshark.Dummy_packets[2]
    convos = <- cstream
    if len(convos) != 1 {
        t.Errorf(fmt.Sprintf("Wrong number of conversations: %d", len(convos)))
    }
}
