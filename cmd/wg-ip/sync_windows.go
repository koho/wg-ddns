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

type DDNSService struct {
}

func (service *DDNSService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	changes <- svc.Status{State: svc.StartPending}

	defer func() {
		changes <- svc.Status{State: svc.StopPending}
	}()

	go func() {
		for {
			if err := sync(iface); err != nil && debug {
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
	return svc.Run("", &DDNSService{})
}

func sync(name string) error {
	config, err := conf.LoadFromName(name)
	if err != nil {
		return err
	}
	directory, err := conf.RootDirectory(false)
	if err != nil {
		return err
	}
	wg := filepath.Join(filepath.Dir(directory), "wg.exe")
	var args []string
	for _, x := range config.Peers {
		if net.ParseIP(x.Endpoint.Host) != nil {
			continue
		}
		args = append(args, "peer", x.PublicKey.String(), "endpoint", x.Endpoint.String())
	}
	if len(args) == 0 {
		return nil
	}
	if output, err := exec.Command(wg, append([]string{"set", name}, args...)...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	return nil
}
