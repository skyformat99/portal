package portal

import (
	"sync"
	"unsafe"

	"github.com/SentimensRG/ctx"
	radix "github.com/armon/go-radix"
	"github.com/pkg/errors"
)

var addrTable = addrSpace{slots: newSlotTable()}

type boundEndpoint interface {
	Endpoint
	ConnectEndpoint(Endpoint)
}
type slotTable radix.Tree

func newSlotTable() *slotTable { return (*slotTable)(unsafe.Pointer(radix.New())) }

func (s *slotTable) Occupied(slot string) (exists bool) {
	_, exists = s.Get(slot)
	return
}

func (s *slotTable) Insert(slot string, ep boundEndpoint) {
	(*radix.Tree)(unsafe.Pointer(s)).Insert(slot, ep)
}

func (s *slotTable) Get(slot string) (ep boundEndpoint, ok bool) {
	var v interface{}
	if v, ok = (*radix.Tree)(unsafe.Pointer(s)).Get(slot); ok {
		ep = v.(boundEndpoint)
	}
	return
}

func (s *slotTable) Del(slot string) { (*radix.Tree)(unsafe.Pointer(s)).Delete(slot) }

type addrSpace struct {
	sync.RWMutex
	slots *slotTable
}

func (a *addrSpace) Assign(addr string, ep boundEndpoint) (err error) {
	a.Lock()
	defer a.Unlock()

	if a.slots.Occupied(addr) {
		err = errors.New("address in use")
	} else {
		a.slots.Insert(addr, ep)
		ctx.Defer(ep, a.releaseSlot(addr))
	}

	return
}

func (a *addrSpace) Lookup(addr string) (ep boundEndpoint, err error) {
	a.RLock()
	defer a.RUnlock()

	var ok bool
	if ep, ok = a.slots.Get(addr); !ok {
		err = errors.New("unbound address")
	}

	return
}

func (a *addrSpace) releaseSlot(addr string) func() {
	return func() {
		a.Lock()
		a.slots.Del(addr)
		a.Unlock()
	}
}
