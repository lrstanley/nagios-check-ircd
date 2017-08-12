# check-ircd

Nagios utility for monitoring the health of an ircd.

## Table of Contents
- [Installation](#installation)
  - [Ubuntu/Debian](#ubuntudebian)
  - [CentOS/Redhat](#centosredhat)
  - [Manual Install](#manual-install)
  - [Source](#source)
- [Usage](#usage)
- [Example Nagios config](#example-nagios-config)
- [License](#license)

## Installation

Check out the [releases](https://github.com/lrstanley/nagios-check-ircd/releases)
page for prebuilt versions. check-ircd should work on ubuntu/debian,
centos/redhat/fedora, etc. Below are example commands of how you would install
the utility (ensure to replace `${VERSION...}` etc, with the appropriate vars).

### Ubuntu/Debian

```bash
$ wget https://github.com/lrstanley/nagios-check-ircd/releases/download/${VERSION}/check-ircd_${VERSION_OS_ARCH}.deb
$ dpkg -i check-ircd_${VERSION_OS_ARCH}.deb
```

### CentOS/Redhat

```bash
$ yum localinstall https://github.com/lrstanley/nagios-check-ircd/releases/download/${VERSION}/check-ircd_${VERSION_OS_ARCH}.rpm
```

### Manual Install

```bash
$ wget https://github.com/lrstanley/nagios-check-ircd/releases/download/${VERSION}/check-ircd_${VERSION_OS_ARCH}.tar.gz
$ tar -C /usr/bin/ -xzvf check-ircd_${VERSION_OS_ARCH}.tar.gz check-ircd
$ chmod +x /usr/bin/check-ircd
```

### Source

If you need a specific version, feel free to compile from source (you must
install [Go](https://golang.org/doc/install) first):

```
$ git clone https://github.com/lrstanley/nagios-check-ircd.git
$ cd nagios-check-ircd
$ make help
$ make build
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

## Exampe Nagios Config

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

## License

```
LICENSE: The MIT License (MIT)
Copyright (c) 2017 Liam Stanley <me@liamstanley.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
