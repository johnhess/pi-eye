package hists

import (
    "fmt"
    "os"
    "strconv"
    "pi-eye/internal/forweb"
    "pi-eye/internal/tshark"
)

type TrafChunk struct {
    Timestamp int64
    Count int
}

type ConversationHist struct {
    Source string
    Destination string
    Traffic []TrafChunk
}

func pkts2hist(pstream <- chan tshark.Packet, hstream chan <- []ConversationHist, delta int64, windows int) {
    go func() {
        // ip_addr: position in dh
        devices := make(map[string]int)
        dh := make([]ConversationHist, 0)
        var offset int64 = -1
        var lastsent int64 = -1
        pkts := 0

        for {
            select {
            case packet := <- pstream:
                pkts += 1
                fmt.Println("processing packet", pkts)
                tm, err := strconv.ParseInt(packet.Timestamp, 10, 64)
                if err != nil {
                    panic(err)
                }
                if offset == -1 {
                    offset = tm
                    for index, convo := range dh {
                        chunk := TrafChunk{offset, 0}
                        dh[index].Traffic = append(convo.Traffic, chunk)
                    }
                } else if tm >= offset + delta {
                    for {
                        offset = offset + delta
                        hstream <- dh
                        for index, convo := range dh {
                            chunk := TrafChunk{offset, 0}
                            dh[index].Traffic = append(convo.Traffic, chunk)
                            traflen := len(dh[index].Traffic)
                            if traflen > windows {
                                dh[index].Traffic = dh[index].Traffic[traflen-windows:]
                            }
                        }
                        if tm <= offset + delta {break}
                    }
                }
                src, dest := packet.Fromto()
                convo := src + ":" + dest
                if _, ok := devices[convo]; !ok {
                    // device not yet in map or array
                    newdh := ConversationHist{src, dest, []TrafChunk{TrafChunk{offset, 0}}}
                    dh = append(dh, newdh)
                    devices[convo] = len(dh) - 1
                }
                dtraf := dh[devices[convo]].Traffic
                dtraf[len(dtraf) - 1].Count += packet.Size()
                if offset > lastsent {
                    fmt.Println("exporting histogram")
                    hstream <- dh
                    lastsent = offset
                }
            }
        }
    }()
}

func Hists() {

    var resms int64 = 1000
    hist_windows := 250

    pstream := make(chan tshark.Packet, 1000)
    hstream := make(chan []ConversationHist)

    tshark.Si2pkts(pstream)
    pkts2hist(pstream, hstream, resms, hist_windows)

    // could be part of pkts2hist, and just write to disk
    for {
        var lastts int64 = 0
        select {
        case hist := <- hstream:
            var newts int64 = 0
            if len(hist) > 0 {
                newts = hist[0].Traffic[len(hist[0].Traffic) - 1].Timestamp
            }
            if newts != lastts {
                forweb.Save(hist, os.Getenv("GOPATH") + "/src/pi-eye/web/visualization/hist.json")
                lastts = newts
            }
        }
    }
}