// Package main implements knockr, a port-knocking utility.
//
// # MIT License
//
// # Copyright (c) 2023 Mike Watkins
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultTimeout = time.Second
	defaultDelay   = 100 * time.Millisecond
)

// version is set via -ldflags "-X main.version=..."; binaries built with
// `go install` fall back to module build info.
var version = ""

// knockPayload is sent on udp knocks; connectionless protocols require a
// packet, not just a dial.
var knockPayload = []byte{0xDE, 0xCA, 0xFB, 0xAD}

var validProtocols = map[string]bool{
	"tcp": true, "tcp4": true, "tcp6": true,
	"udp": true, "udp4": true, "udp6": true,
}

type config struct {
	protocol string
	delay    time.Duration
	timeout  time.Duration
	silent   bool
	version  bool
	verify   string
	config   string
}

// usageError marks argument errors, which warrant printing usage; runtime
// errors do not.
type usageError struct{ err error }

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "knockr error: %s\n", err)

		var ue *usageError
		if errors.As(err, &ue) {
			fmt.Fprintln(os.Stderr)
			flag.Usage()
		}
		os.Exit(1)
	}
}

func run() error {
	cfg := &config{
		protocol: "tcp",
		delay:    defaultDelay,
		timeout:  defaultTimeout,
	}
	flag.Usage = usage
	flag.DurationVar(&cfg.delay, "d", cfg.delay, "`delay` between knocks")
	flag.DurationVar(&cfg.timeout, "t", cfg.timeout, "`timeout` for each knock")
	flag.StringVar(&cfg.protocol, "n", cfg.protocol, "`network` protocol (tcp, tcp4, tcp6, udp, udp4, udp6)")
	flag.BoolVar(&cfg.silent, "s", cfg.silent, "silence all but error output")
	flag.BoolVar(&cfg.version, "version", cfg.version, "print version and exit")
	flag.StringVar(&cfg.verify, "verify", cfg.verify, "tcp `ports` to test after knocking (comma-separated); exits non-zero unless all open")
	flag.StringVar(&cfg.config, "c", cfg.config, "profile `file` (default ~/.ssh/knockr.conf)")
	flag.Parse()

	if cfg.version {
		fmt.Println("knockr", versionString())
		return nil
	}

	// explicit records flags set on the command line; these take precedence
	// over profile values.
	explicit := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })

	var (
		host  string
		ports []int
		err   error
	)
	switch flag.NArg() {
	case 1:
		host, ports, err = fromProfile(cfg, flag.Arg(0), explicit)
		if err != nil {
			return err
		}
	case 2:
		host = flag.Arg(0)
		ports, err = parsePorts(flag.Arg(1))
		if err != nil {
			return &usageError{err}
		}
	default:
		return &usageError{errors.New("expected a profile name, or hostname-or-address and port1,port2")}
	}

	if !validProtocols[cfg.protocol] {
		return &usageError{fmt.Errorf("invalid protocol %q", cfg.protocol)}
	}

	var verifyPorts []int
	if cfg.verify != "" {
		verifyPorts, err = parsePorts(cfg.verify)
		if err != nil {
			err = fmt.Errorf("verify: %w", err)
			// a mistyped flag warrants usage; a bad profile value does not
			if explicit["verify"] {
				return &usageError{err}
			}

			return err
		}
	}

	host, err = resolveHost(host, cfg.protocol)
	if err != nil {
		return err
	}

	portknock(cfg, os.Stdout, host, ports)

	if len(verifyPorts) > 0 {
		return verify(cfg, os.Stdout, host, verifyPorts)
	}

	return nil
}

// verify tests that each verify port is reachable over tcp after the knock
// sequence, matching the knock protocol's address family. Services unlocked
// by knocking (ssh and the like) listen on tcp; udp services such as
// WireGuard cannot be probed meaningfully, so pair them with tcp verify
// ports opened by the same rule.
func verify(cfg *config, out io.Writer, host string, ports []int) error {
	proto := "tcp"
	switch {
	case strings.HasSuffix(cfg.protocol, "4"):
		proto = "tcp4"
	case strings.HasSuffix(cfg.protocol, "6"):
		proto = "tcp6"
	}

	logger := log.New(out, "", log.LstdFlags)
	var failed []string
	for _, port := range ports {
		time.Sleep(cfg.delay)
		address := net.JoinHostPort(host, strconv.Itoa(port))
		result := knock(proto, address, cfg.timeout)
		// knock-oriented annotation does not apply to a verify probe
		result = strings.TrimSuffix(result, " (knock delivered)")

		if !cfg.silent {
			logger.Printf("%s %5d verify: %s", host, port, result)
		}
		if result != "open" {
			failed = append(failed, fmt.Sprintf("port %d %s", port, result))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("verification failed: %s %s", host, strings.Join(failed, "; "))
	}

	return nil
}

func usage() {
	w := flag.CommandLine.Output()
	_, _ = fmt.Fprintf(w, "USAGE\n\nknockr [options] hostname-or-address port1,port2...\nknockr [options] profile-name\n\nOPTIONS\n\n")
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(w, `
EXAMPLES

  Knock using default options:
    knockr somehost.example.com 1234,8923,1233

  Using udp protocol and 50ms delay between knocks:
    knockr -n udp -d 50ms 23.192.228.80 51213,8327,183,420

  Knock then confirm port 22 opened:
    knockr -verify 22 somehost.example.com 1234,8923,1233

  Knock using the "router" profile from ~/.ssh/knockr.conf:
    knockr router

PROFILE FILE (~/.ssh/knockr.conf, must be chmod 600)

  profile router
      host 203.0.113.1
      ports 1234,8923,1233
      protocol udp
      verify 22

`)
}

// versionString reports the ldflags-injected version, falling back to module
// build info for `go install` builds.
func versionString() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
		return bi.Main.Version
	}
	return "(unknown)"
}

// parsePorts converts a comma-separated list into validated port numbers.
func parsePorts(s string) ([]int, error) {
	fields := strings.Split(s, ",")
	ports := make([]int, 0, len(fields))
	for _, f := range fields {
		p, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil {
			return nil, fmt.Errorf("invalid port %q", strings.TrimSpace(f))
		}
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("port %d; allowable ports are 1 - 65535", p)
		}
		ports = append(ports, p)
	}

	return ports, nil
}

// resolveHost pins a single IP address for the entire knock sequence. Go does
// not cache DNS lookups; without pinning, a multi-record host could
// round-robin between knocks and break the sequence. The address family is
// matched to protocol when tcp4/tcp6/udp4/udp6 is specified.
func resolveHost(host, protocol string) (string, error) {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String(), nil
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}

	want := 0
	switch {
	case strings.HasSuffix(protocol, "4"):
		want = 4
	case strings.HasSuffix(protocol, "6"):
		want = 6
	}

	for _, a := range addrs {
		ip := net.ParseIP(a)
		if ip == nil {
			continue
		}
		switch want {
		case 4:
			if ip.To4() != nil {
				return a, nil
			}
		case 6:
			if ip.To4() == nil {
				return a, nil
			}
		default:
			return a, nil
		}
	}

	return "", fmt.Errorf("no %s address found for %s", protocol, host)
}

// portknock knocks each port in sequence, waiting cfg.delay between knocks.
func portknock(cfg *config, out io.Writer, host string, ports []int) {
	logger := log.New(out, "", log.LstdFlags)
	for i, port := range ports {
		if i > 0 {
			time.Sleep(cfg.delay)
		}

		address := net.JoinHostPort(host, strconv.Itoa(port))
		result := knock(cfg.protocol, address, cfg.timeout)

		if !cfg.silent {
			logger.Printf("%s %5d %s", host, port, result)
		}
	}
}

// knock dials address once and describes the outcome. For port-knocking, a
// refused connection still delivers the knock; a timeout is the expected
// result on a stealth-dropped port.
func knock(protocol, address string, timeout time.Duration) string {
	conn, err := net.DialTimeout(protocol, address, timeout)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return "refused (knock delivered)"
		}
		var nerr net.Error
		if errors.As(err, &nerr) && nerr.Timeout() {
			return "no response (timeout)"
		}

		return err.Error()
	}
	defer func() { _ = conn.Close() }()

	if strings.HasPrefix(protocol, "udp") {
		if _, err := conn.Write(knockPayload); err != nil {
			return err.Error()
		}

		return "udp packet sent"
	}

	return "open"
}
