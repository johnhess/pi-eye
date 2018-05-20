package convos

import (
    "os"
    "pi-eye/internal/forweb"
    "pi-eye/internal/tshark"
)

type Conversation struct {
    Source string
    Destination string
}

func pkts2convos(pstream <- chan tshark.Packet, cstream chan <- map[string]Conversation) {
    go func () {
        convos := make(map[string]Conversation)
        ct := 0
        for {
            select {
            case packet := <- pstream:
                from, to := packet.Fromto();
                key := from + ":" + to
                convos[key] = Conversation{from, to}
                if ct < len(convos) {
                    // Don't send map out by reference -- we'll want to
                    // concurrently add to it.
                    out := make(map[string]Conversation)
                    for key, value := range convos {
                        out[key] = value
                    }
                    ct = len(convos)
                    cstream <- out
                }
            }
        }
    }() 
}

/*
 * Write all pairwise conversations observed to a file
 */
func Convos() {
    pstream := make(chan tshark.Packet, 1000)
    cstream := make(chan map[string]Conversation)

    tshark.Si2pkts(pstream)
    pkts2convos(pstream, cstream)

    for {
        select {
        case convos := <- cstream:
            forweb.Save(convos, os.Getenv("GOPATH") + "/src/pi-eye/web/visualization/convos.json")
        }
    }
}