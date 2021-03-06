package net

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
)

var IPv4RouteParser = regexp.MustCompile("^([\\d\\.]+)\\s+([\\d\\.]+)\\s+([\\d\\.]+)\\s+([A-Z]+)\\s+\\d+\\s+\\d+\\s+\\d+\\s+(.+)$")
var IPv4RouteTokens = 6

func FindInterface(name string) (*Endpoint, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		mac := iface.HardwareAddr.String()
		addrs, err := iface.Addrs()
		// is interface active?
		if err == nil && len(addrs) > 0 {
			if (name == "" && iface.Name != "lo") || iface.Name == name {
				var e *Endpoint = nil
				// For every address of the interface.
				for _, addr := range addrs {
					ip := addr.String()
					// Make sure this is an IPv4 address.
					if m, _ := regexp.MatchString("^[0-9\\.]+/?\\d*$", ip); m == true {
						if strings.IndexRune(ip, '/') == -1 {
							// plain ip
							e = NewEndpointNoResolve(ip, mac, iface.Name, 0)
						} else {
							// ip/bits
							parts := strings.Split(ip, "/")
							ip_part := parts[0]
							bits, err := strconv.Atoi(parts[1])
							if err == nil {
								e = NewEndpointNoResolve(ip_part, mac, iface.Name, uint32(bits))
							}
						}
					}
				}

				if e != nil {
					return e, nil
				}
			}
		}
	}

	if name == "" {
		return nil, fmt.Errorf("Could not find default network interface.")
	} else {
		return nil, fmt.Errorf("Could not find interface '%s'.", name)
	}
}

func FindGateway(iface *Endpoint) (*Endpoint, error) {
	output, err := core.Exec("route", []string{"-n", "-A", "inet"})
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(output, "\n") {
		m := IPv4RouteParser.FindStringSubmatch(line)
		if len(m) == IPv4RouteTokens {
			// destination := m[1]
			// mask := m[3]
			flags := m[4]
			ifname := m[5]

			if ifname == iface.Name() && flags == "UG" {
				gateway := m[2]
				// log.Debugf("Gateway ip is %s", gateway)
				if gateway == iface.IpAddress {
					// log.Debug("Gateway == Interface")
					return iface, nil
				} else {
					// we have the address, now we need its mac
					mac, err := ArpLookup(iface.Name(), gateway, false)
					if err == nil {
						// log.Debugf("Gateway mac is %s", mac)
						return NewEndpoint(gateway, mac), nil
					} else {
						log.Error(err)
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("Could not detect the gateway.")
}
