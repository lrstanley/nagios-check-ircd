package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"net"

	flags "github.com/jessevdk/go-flags"
	"github.com/lrstanley/girc"
)

type Config struct {
	Host string `short:"H" long:"host" description:"irc server hostname or address" required:"true"`
	Port int    `short:"p" long:"port" description:"irc server port" default:"6667"`
	Nick string `short:"n" long:"nick" description:"nickname to use" default:"nagios-check"`
	User string `short:"u" long:"user" description:"username (ident) to use" default:"nagios"`
	V4   bool   `short:"4" description:"connect to the irc server via IPv4"`
	V6   bool   `short:"6" description:"connect to the irc server via IPv6"`
	TLS  struct {
		Use       bool `long:"use" description:"enable tls checks"`
		ValidCert bool `long:"check-cert" description:"if TLS certificate should be verified"`
	} `group:"TLS Options" namespace:"tls"`
	Timeout time.Duration `short:"t" long:"timeout" description:"time before the connection attempt should be abandoned" default:"30s"`
	Debug   bool          `short:"d" long:"debug" description:"enable debug output"`
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

	if conf.V4 && !conf.V6 {
		conf.Host = ""
		ips, err := net.LookupIP(originHost)
		if err != nil {
			fmt.Fprintln(os.Stderr, "CRITICAL: "+err.Error())
			os.Exit(1)
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
		ips, err := net.LookupIP(originHost)
		if err != nil {
			fmt.Fprintln(os.Stderr, "CRITICAL: "+err.Error())
			os.Exit(1)
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
		fmt.Fprintln(os.Stderr, "CRITICAL: no record for "+originHost)
		os.Exit(1)
	}

	if conf.TLS.Use {
		err = check(conf.Nick, conf.User, conf.Host, conf.Port, &tls.Config{InsecureSkipVerify: !conf.TLS.ValidCert})
	} else {
		err = check(conf.Nick, conf.User, conf.Host, conf.Port, nil)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "CRITICAL: "+err.Error())
		os.Exit(1)
	} else {
		fmt.Println("SUCCESS")
	}

	os.Exit(0)
}

func check(nick, user, host string, port int, tlsConfig *tls.Config) error {
	done := make(chan bool, 1)
	errs := make(chan error, 1)

	ircConf := girc.Config{
		Server:    host,
		Port:      port,
		Nick:      nick,
		User:      user,
		SSL:       tlsConfig != nil,
		TLSConfig: tlsConfig,
	}

	if conf.Debug {
		ircConf.Debug = os.Stdout
	}

	client := girc.New(ircConf)

	client.Handlers.AddBg(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		client.Close()
		done <- true
	})

	go func() {
		if err := client.Connect(); err != nil {
			errs <- err
		}

		done <- true
	}()

	select {
	case err := <-errs:
		return err
	case <-done:
		return nil
	case <-time.After(conf.Timeout):
		return fmt.Errorf("TIMEOUT %ds", int(conf.Timeout.Seconds()))
	}
}

func ipsToString(ips []net.IP) (out []string) {
	for _, ip := range ips {
		out = append(out, ip.String())
	}

	return out
}
