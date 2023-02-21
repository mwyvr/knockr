# knockr

`knockr` is a [port-knocking](https://en.wikipedia.org/wiki/Port_knocking)
utility more convenient to use than `nmap` or `netcat` or other general purpose
tools. Written in Go to be built on Linux, BSD/Unix, Windows and Mac, `knockr`
is a single binary. Example usage:

	knockr -v -p 1234 -p 8923 -p 1233 my.host.name

## Installation

Most Linux / Windows / Mac:

    go install github.com/solutionroute/knockr

Linux distributions not based on `glibc` such as Alpine Linux or
Void Linux (`musl` variant only):

    # clone the package and build on your system (any system) 
    git clone https://github.com/solutionroute/knockr.git
    cd knockr
    CGO_ENABLED=0 go build

This will build a statically linked version you can use on any Linux distribution.

## Usage

Speak to your network administrator to discover the ports and order required;
typically two, three or even more ports will form the knocking sequence.

    -d duration
            delay between knocks (default 500ms)
    -n string
            network protocol (default "tcp")
    -p value
            one or more ports to knock on
    -t duration
            timeout for each knock (default 50ms)
    -v	verbose: report on each step

    Example:

    # in verbose mode, knock on three ports:
    knockr -v -p 1234 -p 8923 -p 1233 my.host.name

You may choose to include your destination port as the last port in the chain;
doing so with the `-v` (verbose) option will inform whether the knocking
operation was successful.

Default timeout and wait periods should be sufficient for most use cases.

## What is port-knocking?

Port-knocking is a network access method that opens normally closed ports on
a router or host when a specific sequence of ports has received a connection
attempt, usually within a specified and short period of time.

A network access device like a router will typically be configured such that
the target port (not necessarily specified in the port-knocking requests) are
only opened to the IP address issuing the correct knock sequence, further
improving security and resiliency to exploit.

Port-knocking can be configured in many commercial router operating systems and
even some that are accessible to technical consumers such as [Mikrotik RouterOS
devices](https://help.mikrotik.com/docs/display/ROS/Port+knocking).

