package main

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	dbg "github.com/nyaosorg/go-windows-dbg"
	"golang.org/x/sys/windows/svc"
	"golang.zx2c4.com/wireguard/windows/conf"
)

type SVCBService struct {
}

func (service *SVCBService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	changes <- svc.Status{State: svc.StartPending}

	defer func() {
		changes <- svc.Status{State: svc.StopPending}
	}()

	go func() {
		for {
			if err := sync(iface, dnsServer); err != nil && debug {
				dbg.Println(err)
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				return
			case svc.Interrogate:
				changes <- c.CurrentStatus
			}
		}
	}
}

func run() error {
	return svc.Run("", &SVCBService{})
}

func sync(name string, dnsServer string) error {
	config, err := conf.LoadFromName(name)
	if err != nil {
		return err
	}
	var changed bool
	peers := make(map[string]*net.UDPAddr)
	for i, x := range config.Peers {
		if net.ParseIP(x.Endpoint.Host) != nil {
			continue
		}
		addr, err := GetRecord(x.Endpoint.Host, dnsServer)
		if err != nil {
			continue
		}
		peers[x.PublicKey.String()] = addr
		if x.Endpoint.Port != uint16(addr.Port) {
			config.Peers[i].Endpoint.Port = uint16(addr.Port)
			changed = true
		}
	}
	if err = compareAndUpdate(config.Name, peers); err != nil {
		return err
	}
	if changed {
		return config.Save(true)
	}
	return nil
}

func compareAndUpdate(name string, peers map[string]*net.UDPAddr) error {
	directory, err := conf.RootDirectory(false)
	if err != nil {
		return err
	}
	wg := filepath.Join(filepath.Dir(directory), "wg.exe")
	output, err := exec.Command(wg, "showconf", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	config, err := conf.FromWgQuickWithUnknownEncoding(string(output), name)
	if err != nil {
		return err
	}
	var args []string
	for _, peer := range config.Peers {
		addr, err := net.ResolveUDPAddr("udp", peer.Endpoint.String())
		if err != nil {
			continue
		}
		if v, ok := peers[peer.PublicKey.String()]; ok {
			if v.String() != addr.String() {
				args = append(args, "peer", peer.PublicKey.String(), "endpoint", v.String())
			}
		}
	}
	if len(args) == 0 {
		return nil
	}
	if output, err = exec.Command(wg, append([]string{"set", name}, args...)...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	return nil
}
