# knockr

`knockr` is a [port-knocking](https://en.wikipedia.org/wiki/Port_knocking)
utility more convenient to use than `nmap` or `netcat` or other general purpose
tools. Written in Go, the utility is a single binary installable on Linux,
BSD/Unix, Windows and Mac platforms.

## Installation

Most Linux / Windows / Mac:

    go install github.com/solutionroute/knockr@latest

Linux distributions not based on `glibc` such as Alpine Linux or
Void Linux (`musl` variant only):

    # clone the package and build a version without CGO on your system (any system) 
    git clone https://github.com/solutionroute/knockr.git

    CGO_ENABLED=0 go install

This will build a statically linked version you can use on any Linux
distribution.

## Usage

Speak to your network administrator to discover the ports and order required;
typically two, three or even more ports will form the knocking sequence.
Default timeout and wait periods should be sufficient for most use cases.

    Usage: knockr [OPTIONS] address port1,port2...

    -d duration
            delay between knocks (default 500ms)
    -n string
            network protocol (default "tcp")
    -s	silent: suppress all but error output
    -t duration
            timeout for each knock (default 100ms)

    Example:

    # knock on three ports using the default protocol (tcp) and delays
    knockr my.host.name 1234,8923,1233

**Tip**: Include the port you expect to be unlocked as the last port in the
chain; the status output will inform whether the knocking operation was
successful. Example, if intending to access 2200:

    # knock on three ports using the default protocol (tcp) and delays
    knockr my.host.name 1234,8923,1233,2200

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

