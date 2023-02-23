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
	timeoutMS = 1500
	delayMS   = 100
)

type config struct {
	network string
	address string
	ports   []int
	delay   time.Duration
	timeout time.Duration
	silent  bool
}

var version = ""

func main() {
	c := &config{
		network: "tcp",
		delay:   delayMS * time.Millisecond,
		timeout: timeoutMS * time.Millisecond,
	}

	if err := run(c); err != nil {
		if version != "" {
			version = "-" + version
		}

		fmt.Printf("knockr%v error: %s\n\n", version, err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(c *config) error {
	flag.Usage = usage
	flag.DurationVar(&c.delay, "d", c.delay, "delay between knocks")
	flag.DurationVar(&c.timeout, "t", c.timeout, "timeout for each knock")
	flag.StringVar(&c.network, "n", c.network, "network protocol")
	flag.BoolVar(&c.silent, "s", c.silent, "silent: suppress all but error output")
	flag.Parse()

	if len(flag.Args()) != 2 {
		return fmt.Errorf("invalid arguments %v", flag.Args())
	}

	// args are  address port1,port2,port3...
	// no validation of address is being performed
	c.address = flag.Args()[0]

	for _, v := range strings.Split(flag.Args()[1], ",") {
		p, err := strconv.Atoi(v)
		if err != nil {
			return err
		}

		if p < 1 || p > 65535 {
			return fmt.Errorf("port %d; allowable ports are 1 - 65535", p)
		}

		c.ports = append(c.ports, p)
	}

	return portknock(c)
}

func usage() {
	fmt.Printf("Usage: knockr [OPTIONS] address port1,port2...\n\n")
	flag.PrintDefaults()
	fmt.Printf(`
Example:

  # knock on three ports using the default protocol (tcp) and delays
  knockr my.host.name 1234,8923,1233

`)
}

// portknock attempts to make a connection to a port(s); we expect timeout or
// other errors for ports being used as a port-knocking scheme by a router or
// network defense system.
func portknock(cfg *config) error {
	var result string

	// ensure DNS lookup cached or first ports may not be knocked
	_, err := net.LookupHost(cfg.address)
	if err != nil {
		log.Printf("%s: %5s %s", cfg.address, "DNS", err.Error())
	}

	for _, v := range cfg.ports {
		address := fmt.Sprintf("%s:%d", cfg.address, v)

		con, err := net.DialTimeout(cfg.network, address, cfg.timeout)
		if err != nil {
			result = err.Error()
		}

		if err == nil && con != nil {
			result = "open"

			con.Close()
		}

		if !cfg.silent {
			log.Printf("%s: %5d %s", cfg.address, v, result)
		}

		time.Sleep(cfg.delay)
	}

	return nil
}
