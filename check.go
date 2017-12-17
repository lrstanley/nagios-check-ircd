// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/lrstanley/girc"
)

type Config struct {
	Host     string `short:"H" long:"host" description:"irc server hostname or address" required:"true"`
	Port     int    `short:"p" long:"port" description:"irc server port" default:"6667"`
	Nick     string `short:"n" long:"nick" description:"nickname to use" default:"nagios-check"`
	User     string `short:"u" long:"user" description:"username (ident) to use" default:"nagios"`
	Password string `long:"password" description:"irc server password if required"`
	V4       bool   `short:"4" description:"connect to the irc server via IPv4"`
	V6       bool   `short:"6" description:"connect to the irc server via IPv6"`
	TLS      struct {
		Use       bool          `long:"use" description:"enable tls checks"`
		ValidCert bool          `long:"check-cert" description:"if TLS certificate should be verified"`
		MinExpire time.Duration `long:"min-expire" description:"minimum time allowed before warning of an expiring certificate"`
	} `group:"TLS Options" namespace:"tls"`
	RequireRegistration bool          `long:"require-registration" description:"Consider it successful ONLY if we receive RPL_WELCOME"`
	Timeout             time.Duration `short:"t" long:"timeout" description:"time before the connection attempt should be abandoned" default:"30s"`
	Debug               bool          `short:"d" long:"debug" description:"enable debug output"`
}

var conf Config

var debug = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	_, err := flags.Parse(&conf)
	if err != nil {
		if FlagErr, ok := err.(*flags.Error); ok && FlagErr.Type == flags.ErrHelp {
			os.Exit(0)
		}

		// go-flags should print to stderr/stdout as necessary, so we won't.
		os.Exit(1)
	}

	if conf.Debug {
		debug.SetOutput(os.Stdout)
	}

	originHost := conf.Host

	var ips []net.IP

	if conf.V4 && !conf.V6 {
		conf.Host = ""
		ips, err = net.LookupIP(originHost)
		if err != nil {
			fmt.Println("CRITICAL: " + err.Error())
			os.Exit(2)
		}

		debug.Printf("resolved ips: %#v", ipsToString(ips))

		if err == nil {
			for _, ip := range ips {
				if ip.To4() != nil {
					conf.Host = ip.String()
					break
				}
			}
		}
	}

	if conf.V6 && !conf.V4 {
		conf.Host = ""
		ips, err = net.LookupIP(originHost)
		if err != nil {
			fmt.Println("CRITICAL: " + err.Error())
			os.Exit(2)
		}

		debug.Printf("resolved ips: %#v", ipsToString(ips))

		if err == nil {
			for _, ip := range ips {
				if ip.To4() == nil {
					conf.Host = fmt.Sprintf("[%s]", ip.String())
					break
				}
			}
		}
	}

	if conf.Host == "" {
		fmt.Println("CRITICAL: no record for " + originHost)
		os.Exit(2)
	}

	var extra string

	if conf.TLS.Use {
		extra, err = check(&tls.Config{ServerName: originHost, InsecureSkipVerify: !conf.TLS.ValidCert})
	} else {
		extra, err = check(nil)
	}

	if err != nil {
		if strings.Contains(err.Error(), "tls cert expires in") {
			fmt.Println("WARNING: " + err.Error())
			os.Exit(1)
		}

		fmt.Println("CRITICAL: " + err.Error())
		os.Exit(2)
	}

	if extra != "" {
		extra = " " + extra
	}

	fmt.Println("OK" + extra)
	os.Exit(0)
}

func check(tlsConfig *tls.Config) (extra string, err error) {
	done := make(chan bool, 1)
	errs := make(chan error, 10)

	ircConf := girc.Config{
		Server:     conf.Host,
		ServerPass: conf.Password,
		Port:       conf.Port,
		Nick:       conf.Nick,
		User:       conf.User,
		SSL:        tlsConfig != nil,
		TLSConfig:  tlsConfig,
	}

	if conf.Debug {
		ircConf.Debug = os.Stdout
	}

	client := girc.New(ircConf)

	var eventCount int
	event := girc.ALL_EVENTS
	if conf.RequireRegistration {
		event = girc.CONNECTED
	}

	client.Handlers.Add(girc.ALL_EVENTS, func(c *girc.Client, e girc.Event) {
		eventCount++
	})

	client.Handlers.AddBg(event, func(c *girc.Client, e girc.Event) {
		if conf.TLS.Use && conf.TLS.MinExpire > 0 {
			cs, err := c.TLSConnectionState()
			if err != nil {
				errs <- err
				return
			}

			var expires time.Duration

			if expires, err = checkCertExpire(cs); err != nil {
				errs <- err
				return
			}

			extra += fmt.Sprintf("(cert expires in %s)", expires.Truncate(time.Hour))
		}

		client.Close()
		done <- true
	})

	go func() {
		if err := client.Connect(); err != nil {
			errs <- err
		}

		done <- true
	}()

	defer client.Close()

	select {
	case err := <-errs:
		return extra, err
	case <-done:
		return extra, nil
	case <-time.After(conf.Timeout):
		if conf.RequireRegistration && eventCount > 0 {
			return extra, fmt.Errorf("REGISTRATION TIMEOUT %ds (%d events)", int(conf.Timeout.Seconds()), eventCount)
		}
		return extra, fmt.Errorf("TIMEOUT %ds", int(conf.Timeout.Seconds()))
	}
}

func ipsToString(ips []net.IP) (out []string) {
	for _, ip := range ips {
		out = append(out, ip.String())
	}

	return out
}

func checkCertExpire(cs *tls.ConnectionState) (time.Duration, error) {
	var newest time.Duration
	for _, chain := range cs.VerifiedChains {
		for _, cert := range chain {
			expires := cert.NotAfter.Sub(time.Now())
			if newest > expires || newest == time.Duration(0) {
				newest = expires
			}

			if expires < conf.TLS.MinExpire {
				return expires.Truncate(time.Hour), fmt.Errorf("tls cert expires in %s", expires.Truncate(time.Hour))
			}
		}
	}

	return newest, nil
}
