// Package main implements knockr, a port-knocking utility. What is
// port-knocking?
//
// Port-knocking is a network access method that opens normally closed ports on
// a router or host when a specific sequence of ports has received a connection
// attempt, usually within a specified and short period of time. The ports can
// be opened only to the IP address issuing the correct knock sequence, further
// improving security and resiliency to exploit.
//
// Port-knocking can be configured in many commercial router operating systems
// and even some that are accessible to technical consumers such as Mikrotik
// RouterOS.
//
// Example:
//
//	knockr -v -p 1234 -p 8923 -p 1233 my.host.name
//
// Default timeouts and waits should be sufficient for most use cases.
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

const delayMS = 50

// TODO provide a toml and env configuration option
type config struct {
	network string
	address string
	ports   intFlags
	delay   time.Duration
	timeout time.Duration
	verbose bool
}

func main() {
	c := &config{
		network: "tcp",
		delay:   10 * delayMS * time.Millisecond, // 0.5s
		timeout: delayMS * time.Millisecond,
	}

	if err := run(c); err != nil {
		fmt.Printf("\n\nError: %s\n", err)
	}
}

func run(c *config) error {
	flag.Usage = usage
	flag.Var(&c.ports, "p", "one or more ports to knock on")
	flag.BoolVar(&c.verbose, "v", c.verbose, "verbose: report on each step")
	// less commonly used
	flag.DurationVar(&c.delay, "d", c.delay, "delay between knocks")
	flag.DurationVar(&c.timeout, "t", c.timeout, "timeout for each knock")
	flag.StringVar(&c.network, "n", c.network, "network protocol")
	flag.Parse()

	if len(c.ports) == 0 {
		fmt.Printf("knockr: missing port(s)\n\n")
		flag.Usage()
	}

	if len(flag.Args()) != 1 {
		fmt.Printf("knockr: missing address\n\n")
		flag.Usage()
	}

	c.address = flag.Args()[0]

	return portknock(c)
}

// portknock attempts to make a connection to a port(s); we expect timeout or
// other errors for ports being used as a port-knocking scheme by a router or
// network defense system.
func portknock(cfg *config) error {
	var result string

	for _, v := range cfg.ports {
		address := fmt.Sprintf("%s:%d", cfg.address, v)

		con, err := net.DialTimeout(cfg.network, address, cfg.timeout)
		if err != nil {
			if os.IsTimeout(err) {
				result = "timeout"
			} else {
				result = "error"
			}
		}

		if con != nil {
			result = "open"

			con.Close()
		}

		if cfg.verbose {
			log.Printf("%s: %5d %s", cfg.address, v, result)
		}

		time.Sleep(cfg.delay)
	}

	return nil
}

// intFlags is an implementation of flags.Value allowing for multiple -p <port>
// flags to be processed
type intFlags []int

// Set converts a string port value into an integer, appending it in order to
// the list of supplied ports. Ports will be knocked in this order.
func (r *intFlags) Set(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d; allowable ports are 1 - 65535", port)
	}

	*r = append(*r, port)

	return nil
}

// String returns port values as a string joined with ","; this is provided to
// meet the flags.Value interface and is not currently utilized.
func (r *intFlags) String() string {
	s := []string{}
	for _, v := range *r {
		s = append(s, fmt.Sprintf("%d", v))
	}

	return strings.Join(s, ",")
}

// usage prints the help text
func usage() {
	fmt.Printf("Usage:\n\n")
	flag.PrintDefaults()
	fmt.Printf(`
Example:

  # in verbose mode, knock on three ports:
  knockr -v -p 1234 -p 8923 -p 1233 my.host.name
  
`)
}
