package main

import (
    "fmt"
    "testing"
)

func TestFail(t *testing.T) {
    if 1 != 2 {
        t.Error(fmt.Sprintf("Wompity Womp"))
    }
}