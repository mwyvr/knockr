package main

import (
	"bytes"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []int
		wantErr bool
	}{
		{"single", "22", []int{22}, false},
		{"multiple", "22,1234,18923", []int{22, 1234, 18923}, false},
		{"spaces", "22, 80 ,443", []int{22, 80, 443}, false},
		{"low bound", "1", []int{1}, false},
		{"high bound", "65535", []int{65535}, false},
		{"zero", "0", nil, true},
		{"too high", "65536", nil, true},
		{"negative", "-1", nil, true},
		{"garbage", "abc", nil, true},
		{"empty element", "22,,80", nil, true},
		{"empty string", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePorts(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePorts(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parsePorts(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveHostLiterals(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		protocol string
		want     string
	}{
		{"ipv4", "127.0.0.1", "tcp", "127.0.0.1"},
		{"ipv6", "::1", "tcp", "::1"},
		{"bracketed ipv6", "[::1]", "tcp", "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveHost(tt.host, tt.protocol)
			if err != nil {
				t.Fatalf("resolveHost(%q, %q) error = %v", tt.host, tt.protocol, err)
			}
			if got != tt.want {
				t.Fatalf("resolveHost(%q, %q) = %q, want %q", tt.host, tt.protocol, got, tt.want)
			}
		})
	}
}

func TestKnockTCPOpen(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	if got := knock("tcp", ln.Addr().String(), time.Second); got != "open" {
		t.Fatalf("knock open port = %q, want %q", got, "open")
	}
}

func TestKnockTCPRefused(t *testing.T) {
	// Listen then close to obtain a port that actively refuses.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := ln.Addr().String()
	_ = ln.Close()

	if got := knock("tcp", address, time.Second); got != "refused (knock delivered)" {
		t.Fatalf("knock closed port = %q, want refusal", got)
	}
}

func TestKnockUDP(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pc.Close() }()

	received := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 16)
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}
		received <- buf[:n]
	}()

	if got := knock("udp", pc.LocalAddr().String(), time.Second); got != "udp packet sent" {
		t.Fatalf("knock udp = %q, want %q", got, "udp packet sent")
	}

	select {
	case payload := <-received:
		if !bytes.Equal(payload, knockPayload) {
			t.Fatalf("received payload % X, want % X", payload, knockPayload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("udp payload not received")
	}
}

func TestPortknockOutput(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config{protocol: "tcp", delay: time.Millisecond, timeout: time.Second}
	var buf bytes.Buffer
	portknock(cfg, &buf, host, []int{port, port})

	out := buf.String()
	if strings.Count(out, "open") != 2 {
		t.Fatalf("expected two open results, got:\n%s", out)
	}
}

func TestPortknockSilent(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config{protocol: "tcp", delay: time.Millisecond, timeout: time.Second, silent: true}
	var buf bytes.Buffer
	portknock(cfg, &buf, host, []int{port})

	if buf.Len() != 0 {
		t.Fatalf("expected no output in silent mode, got:\n%s", buf.String())
	}
}
