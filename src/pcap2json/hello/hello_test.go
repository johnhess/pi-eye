package main

import (
    "testing"
)

func TestTwo(t *testing.T) {
    two := returntwo()
    if two != 2 {
        t.Errorf("returned two but failing anyways")
    }
}