package main

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

func GetRecord(domain string, dnsServer string) (*net.UDPAddr, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeSVCB)
	m.RecursionDesired = true
	r, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		return nil, err
	}
	if len(r.Answer) == 0 {
		return nil, ErrRecordNotFound
	}
	rr, ok := r.Answer[0].(*dns.SVCB)
	if !ok {
		return nil, ErrRecordNotFound
	}
	var ip net.IP
	var port uint16
	var found bool
	for _, v := range rr.Value {
		switch v := v.(type) {
		case *dns.SVCBAlpn:
			for _, alpn := range v.Alpn {
				if alpn == "wg" {
					found = true
					break
				}
			}
		case *dns.SVCBIPv4Hint:
			ip = v.Hint[0]
		case *dns.SVCBPort:
			port = v.Port
		}
	}
	if !found {
		return nil, fmt.Errorf("wg protocol is not found in alpn list")
	}
	if port == 0 {
		return nil, fmt.Errorf("svcb port is not specified")
	}
	if ip == nil {
		if rr.Target == "." {
			m.SetQuestion(dns.Fqdn(rr.Hdr.Name), dns.TypeA)
		} else {
			m.SetQuestion(dns.Fqdn(rr.Target), dns.TypeA)
		}
		r, _, err = c.Exchange(m, dnsServer)
		if err != nil {
			return nil, err
		}
		if len(r.Answer) == 0 {
			return nil, ErrRecordNotFound
		}
		a, ok := r.Answer[0].(*dns.A)
		if !ok {
			return nil, ErrRecordNotFound
		}
		ip = a.A
	}
	return &net.UDPAddr{IP: ip, Port: int(port)}, nil
}
