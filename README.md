# check-irc

Nagios check script for an ircd

-----

## Releases

Check out the [releases](https://github.com/lrstanley/nagios-check-ircd/releases)
page for prebuilt versions. Currently only amd64 linux, and arm is being built.
If you need a specific version, you will need to install go and built it yourself.

## Installation

check-ircd should work on Ubuntu, CentOS, etc. Below are example commands of
how you would install this (ensure to replace `${VERSION...}` etc, with the
appropriate vars):

```
$ wget https://github.com/lrstanley/nagios-check-ircd/releases/download/${VERSION}/check-ircd_${VERSION_OS_ARCH}.tar.gz
$ tar -C /usr/bin/ -xzvf check-ircd_${VERSION_OS_ARCH}.tar.gz check-ircd
```

An example Nagios config:

```
define host {
	use linux-server
	host_name irc1.yourhost.com
	address irc1.yourhost.com
	hostgroups ircd-servers
}

define host {
	use linux-server
	host_name irc2.yourhost.com
	address irc2.yourhost.com
	hostgroups ircd-servers
}

define hostgroup {
    hostgroup_name  ircd-servers
    alias           IRC Servers
}

define service {
	use                  generic-service
	hostgroup_name       ircd-servers
	service_description  IRC
	check_command        check_irc
}

define service {
	use                  generic-service
	hostgroup_name       ircd-servers
	service_description  IRC-TLS
	check_command        check_irc_tls
}

define command {
	command_name check_irc
	command_line /usr/bin/check-ircd -H $HOSTADDRESS$ -4
}

define command {
	command_name check_irc_tls
	command_line /usr/bin/check-ircd -H $HOSTADDRESS$ -4 -p 6697 --tls.use
}
```

## Usage

```
$ ./check-ircd -h
Usage:
  check-ircd [OPTIONS]

Application Options:
  -H, --host=           irc server hostname or address
  -p, --port=           irc server port (default: 6667)
  -n, --nick=           nickname to use (default: nagios-check)
  -u, --user=           username (ident) to use (default: nagios)
      --password=       irc server password if required
  -4                    connect to the irc server via IPv4
  -6                    connect to the irc server via IPv6
  -t, --timeout=        time before the connection attempt should be abandoned (default: 30s)
  -d, --debug           enable debug output

TLS Options:
      --tls.use         enable tls checks
      --tls.check-cert  if TLS certificate should be verified

Help Options:
  -h, --help            Show this help message
```
