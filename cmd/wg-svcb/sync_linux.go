package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

var opt = ini.LoadOptions{KeyValueDelimiters: "=", IgnoreInlineComment: true}

func run() error {
	for {
		if err := sync(iface, dnsServer); err != nil && debug {
			log.Println(err)
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func sync(name string, dnsServer string) error {
	path := fmt.Sprintf("/etc/wireguard/%s.conf", name)
	f, err := ini.LoadSources(opt, path)
	if err != nil {
		return err
	}
	sections, err := f.SectionsByName("Peer")
	if err != nil {
		return err
	}
	var changed bool
	peers := make(map[string]*net.UDPAddr)
	for _, x := range sections {
		endpoint := x.Key("Endpoint")
		host, port, err := net.SplitHostPort(endpoint.String())
		if err != nil {
			continue
		}
		if net.ParseIP(host) != nil {
			continue
		}
		addr, err := GetRecord(host, dnsServer)
		if err != nil {
			continue
		}
		peers[x.Key("PublicKey").String()] = addr
		if newPort := strconv.Itoa(addr.Port); port != newPort {
			endpoint.SetValue(net.JoinHostPort(host, newPort))
			changed = true
		}
	}
	if err = compareAndUpdate(name, peers); err != nil {
		return err
	}
	if changed {
		return f.SaveTo(path)
	}
	return nil
}

func compareAndUpdate(name string, peers map[string]*net.UDPAddr) error {
	output, err := exec.Command("wg", "showconf", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	f, err := ini.LoadSources(opt, output)
	if err != nil {
		return err
	}
	sections, err := f.SectionsByName("Peer")
	if err != nil {
		return err
	}
	var args []string
	for _, peer := range sections {
		endpoint := peer.Key("Endpoint")
		pk := peer.Key("PublicKey")
		addr, err := net.ResolveUDPAddr("udp", endpoint.String())
		if err != nil {
			continue
		}
		if v, ok := peers[pk.String()]; ok {
			if v.String() != addr.String() {
				args = append(args, "peer", pk.String(), "endpoint", v.String())
			}
		}
	}
	if len(args) == 0 {
		return nil
	}
	if output, err = exec.Command("wg", append([]string{"set", name}, args...)...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	return nil
}
