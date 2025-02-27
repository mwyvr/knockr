knockr(1)

# NAME

knockr - port knocking utility for tcp/udp

# SYNOPSIS

*knockr* [options...] hostname-or-ip-address port1,port2...

# DESCRIPTION

*knockr*, in the order of port(s) specified, opens a connection (tcp) or sends a
packet (udp) to one or more ports at a hostname or IP address.

# OPTIONS

*-d* <duration>
	Delay between knocks (default 100ms).

*-n* <protocol>
	Network protocol (tcp, udp).

*-s*
	Silence all but error output.

*-t* <duration>
	Timeout for each knock (default 1s).

# EXAMPLES

Knock using default options:
	knockr somehost.example.com 1234,8923,1233

Using udp protocol and 50ms delay between knocks:
	knockr -n udp -d 50ms 23.192.228.80 51213,8327,183,420

# AUTHORS

knockr was written by Mike Watkins <git@mikewatkins.ca>. For more information or
to contribute to the project, see <https://github.com/mwyvr/knockr>.
