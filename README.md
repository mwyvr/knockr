# knockr

`knockr` is a [port-knocking](https://en.wikipedia.org/wiki/Port_knocking)
utility potentially more convenient to use than general-purpose tools like
`nmap` or `netcat`. Written in Go, the utility is a single binary that is
installable on any platform Go supports, including Linux, BSD/UNIX, Windows,
and Mac.

## Installation

### Via the Go toolchain

Standard:

    go install github.com/mwyvr/knockr@latest

Without CGO:

    CGO_ENABLED=0 go install github.com/mwyvr/knockr@latest

### From source, with manpage

`go install` places only the binary; to also install the man page:

    git clone https://github.com/mwyvr/knockr && cd knockr
    make install                          # to ~/.local (bin, share/man)
    sudo make install PREFIX=/usr/local   # system-wide

Generating the man page requires [scdoc](https://git.sr.ht/~sircmpwn/scdoc),
packaged in most distributions.

### Pre-built binaries

The [releases page](https://github.com/mwyvr/knockr/releases) provides binaries
for various operating systems and architectures.

## Usage

_The default timeout and delay durations should be sufficient for
most use cases._

knockr [options] hostname-or-address port1,port2...
knockr [options] profile-name

OPTIONS

    -c file
      	profile file (default ~/.ssh/knockr.conf)
    -d delay
      	delay between knocks (default 100ms)
    -n network
      	network protocol (tcp, tcp4, tcp6, udp, udp4, udp6) (default "tcp")
    -s	silence all but error output
    -t timeout
      	timeout for each knock (default 1s)
    -verify ports
      	tcp ports to test after knocking (comma-separated);
      	exits non-zero unless all open
    -version
      	print version and exit

EXAMPLES

    Knock using default options:
      knockr somehost.example.com 1234,8923,1233

    Using udp protocol and 50ms delay between knocks:
      knockr -n udp -d 50ms 23.192.228.80 51213,8327,183,420

**Tip**: Include the port(s) you expect to be unlocked as the first and last
port in the chain to observe the port status before and after. For example, if
intending to unlock port 22 (SSH) on a specific host:

    knockr my.host.name 22,1234,18923,1233,22

## Profiles

Frequently used knock sequences can be stored as named profiles and invoked
with a single argument:

    knockr router

Profiles live in `~/.ssh/knockr.conf` (override with `-c`), in an
ssh_config-style format:

    # ~/.ssh/knockr.conf
    profile router
        host 203.0.113.1
        ports 1234,8923,1233
        protocol udp
        delay 150ms
        timeout 2s
        verify 22

All keys except `host` and `ports` are optional.

Flags given on the command line override profile values. `verify` names one
or more tcp ports (comma-separated) to test after the sequence completes —
knockr exits non-zero unless all of them open, making profiles usable directly
in scripts:

    knockr router && rsync ...

Why put the configuration in `~/.ssh`? Knock sequences are secrets.
Configuration under `~/.config` can be accidentally included in public dotfiles
repositories; `~/.ssh` is one directory users routinely treat as private. knockr
refuses (and warns) to read a profile file that is readable by group or others:

    chmod 600 ~/.ssh/knockr.conf

## What is port-knocking?

Port-knocking is a network access method that opens ports that are normally
closed to the outside world, but only when the correct sequence of ports
has been visited and within time frames determined by your network access
configuration.

A host or network protected by port knocking reduces the log burden from
Internet port scanners and should only be seen as another tool to improve
security.

Port-knocking can be configured on hosts ([iptables or knockd](https://wiki.archlinux.org/title/Port_knocking)), and
many routers, including some low-cost, high-functionality devices
accessible to technical consumers such as [Mikrotik RouterOS devices](https://help.mikrotik.com/docs/display/ROS/Port+knocking).

See also: [Wikipedia - port-knocking](https://en.wikipedia.org/wiki/Port_knocking).
