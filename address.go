package portal

import (
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/pkg/errors"
)

var addrTable = addrSpace{slots: make(map[string]boundEndpoint)}

type boundEndpoint interface {
	Endpoint
	ConnectEndpoint(Endpoint)
}

type addrSpace struct {
	sync.RWMutex
	slots map[string]boundEndpoint
}

func (a *addrSpace) Assign(addr string, ep boundEndpoint) (err error) {
	a.Lock()
	defer a.Unlock()

	if _, exists := a.slots[addr]; exists {
		err = errors.New("address in use")
	} else {
		a.slots[addr] = ep
		ctx.Defer(ep, a.releaseSlot(addr))
	}

	return
}

func (a *addrSpace) Lookup(addr string) (ep boundEndpoint, err error) {
	a.RLock()
	defer a.RUnlock()

	var ok bool
	if ep, ok = a.slots[addr]; !ok {
		err = errors.New("unbound address")
	}

	return
}

func (a *addrSpace) releaseSlot(addr string) func() {
	return func() {
		a.Lock()
		delete(a.slots, addr)
		a.Unlock()
	}
}
