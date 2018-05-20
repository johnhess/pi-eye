package tshark

import (
    "testing"
)

func TestHostLeavesLocalIPsUntouched(t *testing.T) {
    clean := simpleHost("192.168.1.106")
    if clean != "192.168.1.106" {t.Error(clean)}
    clean = simpleHost("172.20.1.1")
    if clean != "172.20.1.1" {t.Error(clean)}
}

func TestHostTruncatesSubdomains(t *testing.T) {
    clean := simpleHost("this.com")
    if clean != "this.com" {t.Error(clean)}
    clean = simpleHost("sub.this.com")
    if clean != "this.com" {t.Error(clean)}
}