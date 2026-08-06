knockr(1)

# NAME

knockr - port knocking utility for tcp/udp

# SYNOPSIS

*knockr* [options...] hostname-or-ip-address port1,port2...++
*knockr* [options...] profile-name

# DESCRIPTION

*knockr*, in the order of port(s) specified, opens a connection (tcp) or sends a
packet (udp) to one or more ports at a hostname or IP address.

# OPTIONS

*-c* <file>
	Profile file (default ~/.ssh/knockr.conf).

*-d* <duration>
	Delay between knocks (default 100ms).

*-n* <protocol>
	Network protocol (tcp, tcp4, tcp6, udp, udp4, udp6).

*-s*
	Silence all but error output.

*-t* <duration>
	Timeout for each knock (default 1s).

*-verify* <ports>
	Comma-separated tcp ports to test after knocking; exits non-zero unless
	all open.

*-version*
	Print version and exit.

# PROFILES

Named knock sequences may be stored in an ssh_config-style file, by default
*~/.ssh/knockr.conf*, and invoked as *knockr profile-name*. The file must not
be readable by group or others (chmod 600). *host* and *ports* are required;
*protocol*, *delay*, *timeout*, *silent*, and *verify* are optional.
Command-line flags override profile values.

```
profile router
    host 203.0.113.1
    ports 1234,8923,1233
    protocol udp
    verify 22
```

# EXAMPLES

Knock using default options:
	knockr somehost.example.com 1234,8923,1233

Using udp protocol and 50ms delay between knocks:
	knockr -n udp -d 50ms 23.192.228.80 51213,8327,183,420

Knock the router profile, then confirm its verify port opened:
	knockr router

# AUTHORS

knockr was written by Mike Watkins <git@mikewatkins.ca>. For more information or
to contribute to the project, see <https://github.com/mwyvr/knockr>.
