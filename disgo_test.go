package main

import (
	"reflect"

	"testing"
)

// relevantIpBytes gets a complete v4 address.
func Test_RelevantIpBytes_V4(t *testing.T) {
	out := "127.0.0.1"
	ip, err := relevantIpBytes("127.0.0.1:3000")
	expect(t, ip, out)
	expect(t, err, nil)
}

// relevantIpBytes gets a complete v6 address.
func Test_RelevantIpBytes_V6(t *testing.T) {
	out := "fd0d:e3cc:c5a1:ed75::"
	ip, err := relevantIpBytes("[fd0d:e3cc:c5a1:ed75:e493:3093:f4bd:1529]:3000")
	expect(t, ip, out)
	expect(t, err, nil)

	out = "::"
	ip, err = relevantIpBytes("[::1]:3000")
	expect(t, ip, out)
	expect(t, err, nil)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("expected %v (%v) - got %v (%v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
