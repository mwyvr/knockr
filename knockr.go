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
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	timeoutMS = 1000
	delayMS   = 100
)

type config struct {
	protocol string
	delay    time.Duration
	timeout  time.Duration
	silent   bool
}

var version = "" // injected via ldflags

func main() {

	if err := run(); err != nil {
		if version != "" {
			version = "-" + version
		}
		fmt.Printf("knockr%v error: %s\n\n", version, err)
		flag.Usage()
		os.Exit(1)
	}
}

func run() error {

	cfg := &config{
		protocol: "tcp",
		delay:    delayMS * time.Millisecond,
		timeout:  timeoutMS * time.Millisecond,
	}
	// parse options
	flag.Usage = usage
	flag.DurationVar(&cfg.delay, "d", cfg.delay, "delay between knocks")
	flag.DurationVar(&cfg.timeout, "t", cfg.timeout, "timeout for each knock")
	flag.StringVar(&cfg.protocol, "n", cfg.protocol, "network protocol (tcp, udp)")
	flag.BoolVar(&cfg.silent, "s", cfg.silent, "silence all but error output")
	flag.Parse()

	// parse required args: address port1,port2,port3...
	if len(flag.Args()) != 2 {
		return fmt.Errorf("invalid arguments %v", flag.Args())
	}

	ports := []int{}
	for _, v := range strings.Split(flag.Args()[1], ",") {
		p, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		if p < 1 || p > 65535 {
			return fmt.Errorf("port %d; allowable ports are 1 - 65535", p)
		}
		ports = append(ports, p)
	}

	return portknock(cfg, flag.Args()[0], ports)
}

func usage() {
	fmt.Printf("Usage: knockr [OPTIONS] hostname-or-address port1,port2...\n\n")
	flag.PrintDefaults()
	fmt.Printf(`
Examples:

  # knock on three ports using tcp and other defaults
  knockr my.host.name 1234,8923,1233
  # using udp protocol with a 50ms delay between, knock on three ports
  knockr -n udp -d 50ms 123.123.123.010 8327,183,420

`)
}

// portknock attempts to make a connection (tcp) or send a packet (udp) to one
// or more ports at host.
func portknock(cfg *config, host string, ports []int) error {
	var result string

	// if parseable, host is an ip adddress; otherwise we assume a hostname.
	ip := net.ParseIP(host)
	if ip == nil {
		// hostname: ensure DNS lookup is cached or first ports may not be knocked
		_, err := net.LookupHost(host)
		if err != nil {
			return err
		}
	} else {
		if ip.To4() != nil {
			host = ip.String()
		} else {
			// format ipv6 appropriately for net.DialTimeout
			host = fmt.Sprintf("[%s]", ip.String())
		}
	}

	delay := time.NewTicker(cfg.delay)

	for _, port := range ports {
		address := fmt.Sprintf("%s:%d", host, port)

		con, err := net.DialTimeout(cfg.protocol, address, cfg.timeout)
		if err != nil {
			result = err.Error()
		} else {
			switch cfg.protocol {
			case "tcp":
				result = "open"
			case "udp":
				// no handshake with a connectionless protocol, so send a DECAFBAD packet
				_, err := con.Write([]byte{0xDE, 0xCA, 0xFB, 0xAD})
				if err != nil {
					result = err.Error()
				} else {
					result = "udp packet sent"
				}
			}
		}
		if con != nil {
			con.Close()
		}

		if !cfg.silent {
			log.Printf("%s %5d %s", host, port, result)
		}

		<-delay.C
	}
	return nil
}
