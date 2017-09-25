package portal

import (
	"sync"

	"github.com/pkg/errors"
)

var router = transRouter{lookup: make(map[string]*transport)}

type connector interface {
	GetEndpoint() Endpoint
	Connect(Endpoint)
}

type listener interface {
	Listen() <-chan Endpoint
	Close()
}

type transport struct {
	svr  Endpoint
	chEp chan Endpoint
}

func (t transport) GetEndpoint() Endpoint   { return t.svr }
func (t transport) Connect(ep Endpoint)     { t.chEp <- ep }
func (t transport) Listen() <-chan Endpoint { return t.chEp }
func (t transport) Close()                  { close(t.chEp) }

type transRouter struct {
	sync.RWMutex
	lookup map[string]*transport
}

func (t *transRouter) GetConnector(a string) (c connector, ok bool) {
	t.RLock()
	defer t.RUnlock()

	c, ok = t.lookup[a]
	return
}

func (t *transRouter) GetListener(a string, ep Endpoint) (listener, error) {
	t.Lock()
	defer t.Unlock()

	if _, exists := t.lookup[a]; exists {
		return nil, errors.Errorf("transport exists at %s", a)
	}

	tpt := &transport{svr: ep, chEp: make(chan Endpoint)}
	t.lookup[a] = tpt
	return tpt, nil
}
