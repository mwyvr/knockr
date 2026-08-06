package main

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testConfig = `# knockr profiles
profile router
    host 203.0.113.1
    ports 1234,8923,1233
    protocol udp
    delay 150ms
    timeout 2s
    verify 22

Profile office
    Host office.example.com
    Ports 7000,7001
`

func writeConfig(t *testing.T, content string, perm os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "knockr.conf")
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestLoadProfiles(t *testing.T) {
	path := writeConfig(t, testConfig, 0o600)

	profiles, err := loadProfiles(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	router := profiles["router"]
	want := map[string]string{
		"host": "203.0.113.1", "ports": "1234,8923,1233",
		"protocol": "udp", "delay": "150ms", "timeout": "2s", "verify": "22",
	}
	if !reflect.DeepEqual(router, want) {
		t.Fatalf("router profile = %v, want %v", router, want)
	}

	// keywords are case-insensitive
	if profiles["office"]["host"] != "office.example.com" {
		t.Fatalf("office profile = %v", profiles["office"])
	}
}

func TestLoadProfilesErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"unknown key", "profile a\n\thost h\n\tbogus v\n", "unknown key"},
		{"key before profile", "host h\n", "before any profile"},
		{"duplicate profile", "profile a\nhost h\nprofile a\n", "duplicate profile"},
		{"duplicate key", "profile a\nhost h\nhost h\n", "duplicate key"},
		{"malformed line", "profile a\nhost\n", "expected 'keyword value'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.content, 0o600)
			_, err := loadProfiles(path)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadProfilesRejectsOpenPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not meaningful on windows")
	}

	path := writeConfig(t, testConfig, 0o644)
	_, err := loadProfiles(path)
	if err == nil || !strings.Contains(err.Error(), "too open") {
		t.Fatalf("error = %v, want permission rejection", err)
	}
}

func TestApplyProfile(t *testing.T) {
	p := map[string]string{
		"host": "203.0.113.1", "ports": "1234,8923",
		"protocol": "udp", "delay": "150ms", "timeout": "2s",
		"silent": "true", "verify": "22",
	}

	cfg := &config{protocol: "tcp", delay: defaultDelay, timeout: defaultTimeout}
	host, ports, err := applyProfile(cfg, p, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if host != "203.0.113.1" || !reflect.DeepEqual(ports, []int{1234, 8923}) {
		t.Fatalf("host = %q, ports = %v", host, ports)
	}
	if cfg.protocol != "udp" || cfg.delay != 150*time.Millisecond ||
		cfg.timeout != 2*time.Second || !cfg.silent || cfg.verify != "22" {
		t.Fatalf("profile not applied: %+v", cfg)
	}
}

func TestApplyProfileFlagPrecedence(t *testing.T) {
	p := map[string]string{
		"host": "h", "ports": "22", "protocol": "udp", "delay": "150ms", "verify": "22",
	}

	// -n, -d, and -verify were set explicitly and must win.
	cfg := &config{protocol: "tcp6", delay: time.Second, timeout: defaultTimeout, verify: "8080"}
	explicit := map[string]bool{"n": true, "d": true, "verify": true}
	if _, _, err := applyProfile(cfg, p, explicit); err != nil {
		t.Fatal(err)
	}
	if cfg.protocol != "tcp6" || cfg.delay != time.Second || cfg.verify != "8080" {
		t.Fatalf("explicit flags overridden: %+v", cfg)
	}
}

func TestApplyProfileMissingRequired(t *testing.T) {
	cfg := &config{}
	if _, _, err := applyProfile(cfg, map[string]string{"ports": "22"}, nil); err == nil {
		t.Fatal("expected error for missing host")
	}
	if _, _, err := applyProfile(cfg, map[string]string{"host": "h"}, nil); err == nil {
		t.Fatal("expected error for missing ports")
	}
}

func TestFromProfile(t *testing.T) {
	path := writeConfig(t, testConfig, 0o600)

	cfg := &config{protocol: "tcp", delay: defaultDelay, timeout: defaultTimeout, config: path}
	host, ports, err := fromProfile(cfg, "router", map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if host != "203.0.113.1" || len(ports) != 3 || cfg.protocol != "udp" {
		t.Fatalf("host = %q, ports = %v, cfg = %+v", host, ports, cfg)
	}

	_, _, err = fromProfile(cfg, "nope", map[string]bool{})
	if err == nil || !strings.Contains(err.Error(), "available: office, router") {
		t.Fatalf("error = %v, want profile listing", err)
	}
}

func TestVerify(t *testing.T) {
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
	if err := verify(cfg, &buf, host, []int{port}); err != nil {
		t.Fatalf("verify open port: %v", err)
	}
	if !strings.Contains(buf.String(), "verify: open") {
		t.Fatalf("output = %q", buf.String())
	}

	// closed port must fail verification
	_ = ln.Close()
	if err := verify(cfg, &buf, host, []int{port}); err == nil {
		t.Fatal("expected verification failure on closed port")
	}
}

func TestVerifyMultiplePorts(t *testing.T) {
	open1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = open1.Close() }()
	open2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = open2.Close() }()

	// obtain a closed port by listening then closing
	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_ = closed.Close()

	portOf := func(ln net.Listener) int {
		t.Helper()
		_, s, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		p, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}

		return p
	}
	p1, p2, pc := portOf(open1), portOf(open2), portOf(closed)

	cfg := &config{protocol: "tcp", delay: time.Millisecond, timeout: time.Second}
	var buf bytes.Buffer

	// all open: succeeds, both reported
	if err := verify(cfg, &buf, "127.0.0.1", []int{p1, p2}); err != nil {
		t.Fatalf("verify two open ports: %v", err)
	}
	if strings.Count(buf.String(), "verify: open") != 2 {
		t.Fatalf("output = %q", buf.String())
	}

	// one closed: all ports still probed, error names only the failure
	buf.Reset()
	err = verify(cfg, &buf, "127.0.0.1", []int{p1, pc, p2})
	if err == nil {
		t.Fatal("expected verification failure")
	}
	if !strings.Contains(err.Error(), strconv.Itoa(pc)) {
		t.Fatalf("error %v does not name failed port %d", err, pc)
	}
	if strings.Contains(err.Error(), strconv.Itoa(p1)) || strings.Contains(err.Error(), strconv.Itoa(p2)) {
		t.Fatalf("error %v names an open port", err)
	}
	if strings.Count(buf.String(), "verify: open") != 2 {
		t.Fatalf("closed port stopped probing: %q", buf.String())
	}
}
