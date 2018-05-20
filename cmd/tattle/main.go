package main

import (
    "flag"
    "log"
    "os"
    "runtime/pprof"
    "pi-eye/internal/convos"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {

    // If requested, profile the code
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }

    convos.Convos()
}