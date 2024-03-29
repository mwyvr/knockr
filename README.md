# knockr

`knockr` is a [port-knocking](https://en.wikipedia.org/wiki/Port_knocking)
utility potentially more convenient to use than a general purpose tool like 
`nmap` or `netcat`. Written in Go, the utility is a single binary, installable 
on any platform Go supports including Linux, BSD/Unix, Windows and Mac.

## Installation

### Via the Go tool chain

    go install github.com/mwyvr/knockr@latest

**Linux without** `glibc`: The Go `net` package includes CGO bindings; Linux
distributions not based on `glibc` such as [Alpine
Linux](https://www.alpinelinux.org/), [Chimera
Linux](https://chimera-linux.org/) or [Void Linux](https://voidlinux.org/)
(`musl` variant) can install a statically linked version with:

    CGO_ENABLED=0 go install github.com/mwyvr/knockr@latest

### Other Install Options 

**Pre-built binary for Linux**:

The [releases page](https://github.com/mwyvr/knockr/releases)
provides a link to a non CGO-based binary that will run on various
Linux distributions.

## Usage

*The default timeout and delay durations should be sufficient for
most use cases.*

    Usage: knockr [OPTIONS] address port1,port2...

    -d duration
            delay between knocks (default 100ms)
    -n string
            network protocol (default "tcp")
    -s	silent: suppress all but error output
    -t duration
            timeout for each knock (default 1.5s)

    Example:

    # knock on three ports using the default protocol (tcp) and delays
    knockr my.host.name 1234,8923,1233

**Tip**: Include the port(s) you expect to be unlocked as the first and last
port in the chain to observe status before and after. For example, if intending
to unlock port 22 (ssh) on a specific host:

    # 22 last to demonstrate it has been opened
    knockr my.host.name 1234,18923,1233,22

## What is port-knocking?

Port-knocking is a network access method that opens ports normally left closed
to the outside world, but only when the right sequence of ports has been
visited and within time frames determined by your network access configuration.
That sequence of ports acts as a key.

knockr is the remote side of the solution; a network access device like a
router must be configured.

Port-knocking can be configured on hosts and many routers including some
low-cost, high functionality devices accessible to technical consumers such as
[Mikrotik RouterOS devices](https://help.mikrotik.com/docs/display/ROS/Port+knocking).

Typically the solution will be configured such that the target port (not
necessarily specified in the port-knocking requests) are only opened to the IP
address issuing the correct knock sequence, further improving security and
resiliency to exploit, and reducing port-scanning log burden.

See also: [Wikipedia - port-knocking](https://en.wikipedia.org/wiki/Port_knocking).
