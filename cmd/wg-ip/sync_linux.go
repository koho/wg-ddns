package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"gopkg.in/ini.v1"
)

var opt = ini.LoadOptions{KeyValueDelimiters: "=", IgnoreInlineComment: true}

func run() error {
	for {
		if err := sync(iface); err != nil && debug {
			log.Println(err)
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func sync(name string) error {
	path := fmt.Sprintf("/etc/wireguard/%s.conf", name)
	f, err := ini.LoadSources(opt, path)
	if err != nil {
		return err
	}
	sections, err := f.SectionsByName("Peer")
	if err != nil {
		return err
	}
	var args []string
	for _, x := range sections {
		endpoint := x.Key("Endpoint")
		host, _, err := net.SplitHostPort(endpoint.String())
		if err != nil {
			continue
		}
		if net.ParseIP(host) != nil {
			continue
		}
		args = append(args, "peer", x.Key("PublicKey").String(), "endpoint", endpoint.String())
	}
	if len(args) == 0 {
		return nil
	}
	if output, err := exec.Command("wg", append([]string{"set", name}, args...)...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	return nil
}
