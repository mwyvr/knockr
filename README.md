# knockr

`knockr` is a [port-knocking](https://en.wikipedia.org/wiki/Port_knocking)
utility potentially more convenient to use than general-purpose tools like 
`nmap` or `netcat`. Written in Go, the utility is a single binary that is installable
on any platform Go supports, including Linux, BSD/Unix, Windows, and Mac.

## Installation

### Via the Go toolchain

Standard:

    go install github.com/mwyvr/knockr@latest

Without CGO:

    CGO_ENABLED=0 go install github.com/mwyvr/knockr@latest

### Pre-built binaries

The [releases page](https://github.com/mwyvr/knockr/releases) provides binaries
for various operating systems and architectures.

## Usage

*The default timeout and delay durations should be sufficient for
most use cases.*

    Usage: knockr [OPTIONS] hostname-or-address port1,port2...

    -d duration
            delay between knocks (default 100ms)
    -n string
            network protocol (tcp, udp) (default "tcp")
    -s	silence all but error output
    -t duration
            timeout for each knock (default 1s)

    Examples:

      # knock on three ports using tcp and other defaults
      knockr my.host.name 1234,8923,1233
      # using udp protocol with a 50ms delay between, knock on three ports
      knockr -n udp -d 50ms 123.123.123.010 8327,183,420

**Tip**: Include the port(s) you expect to be unlocked as the first and last
port in the chain to observe the port status before and after. For example, if
intending to unlock port 22 (SSH) on a specific host:

    knockr my.host.name 22,1234,18923,1233,22

## What is port-knocking?

Port-knocking is a network access method that opens ports that are normally closed
to the outside world, but only when the correct sequence of ports has been
visited and within time frames determined by your network access configuration.

A host or network protected by port knocking reduces the log burden from
Internet port scanners can be another tool to improve security further.

Port-knocking can be configured on hosts ([iptables or
knockd](https://wiki.archlinux.org/title/Port_knocking)), and
many routers, including some low-cost, high-functionality devices
accessible to technical consumers such as [Mikrotik RouterOS
devices](https://help.mikrotik.com/docs/display/ROS/Port+knocking).

See also: [Wikipedia - port-knocking](https://en.wikipedia.org/wiki/Port_knocking).
