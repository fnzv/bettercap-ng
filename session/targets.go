package session

import (
	"fmt"
	"sort"
	"sync"

	"github.com/op/go-logging"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/net"
)

var log = logging.MustGetLogger("mitm")

type Targets struct {
	Interface *net.Endpoint
	Gateway   *net.Endpoint
	Targets   map[string]*net.Endpoint
	lock      sync.Mutex
}

func NewTargets(iface, gateway *net.Endpoint) *Targets {
	return &Targets{
		Interface: iface,
		Gateway:   gateway,
		Targets:   make(map[string]*net.Endpoint),
	}
}

func (tp *Targets) Remove(ip, mac string) {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	if e, found := tp.Targets[mac]; found {
		log.Infof("[%slost%s] %s\n", core.RED, core.RESET, e)
		delete(tp.Targets, mac)
		return
	}
}

func (tp *Targets) shouldIgnore(ip string) bool {
	return (ip == tp.Interface.IpAddress || ip == tp.Gateway.IpAddress)
}

func (tp *Targets) Has(ip string) bool {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	for _, e := range tp.Targets {
		if e.IpAddress == ip {
			return true
		}
	}

	return false
}

func (tp *Targets) AddIfNotExist(ip, mac string) *net.Endpoint {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	if tp.shouldIgnore(ip) {
		return nil
	}

	if t, found := tp.Targets[mac]; found {
		return t
	}

	e := net.NewEndpoint(ip, mac)
	log.Infof("[%snew%s] %s\n", core.GREEN, core.RESET, e)
	tp.Targets[mac] = e

	return nil
}

type tSorter []*net.Endpoint

func (a tSorter) Len() int           { return len(a) }
func (a tSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a tSorter) Less(i, j int) bool { return a[i].IpAddressUint32 < a[j].IpAddressUint32 }

func (tp *Targets) Dump() {
	tp.lock.Lock()
	defer tp.lock.Unlock()

	fmt.Println()
	fmt.Printf("  " + core.GREEN + "interface" + core.RESET + "\n\n")
	fmt.Printf("    " + tp.Interface.String() + "\n")
	fmt.Println()
	fmt.Printf("  " + core.GREEN + "gateway" + core.RESET + "\n\n")
	fmt.Printf("    " + tp.Gateway.String() + "\n")

	if len(tp.Targets) > 0 {
		fmt.Println()
		fmt.Printf("  " + core.GREEN + "hosts" + core.RESET + "\n\n")
		targets := make([]*net.Endpoint, 0, len(tp.Targets))
		for _, t := range tp.Targets {
			targets = append(targets, t)
		}

		sort.Sort(tSorter(targets))

		for _, t := range targets {
			fmt.Println("    " + t.String())
		}
	}

	fmt.Println()
}
