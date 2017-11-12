package portal

import (
	"context"
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/pkg/errors"
)

type bindKey uint

const (
	keySvrEndpt bindKey = iota
	keyListenChan
	keyBindAddr
)

var transport = trans{lookup: make(map[string]*bindCtx)}

type (
	transporter interface {
		Connect(addr string) error
		Bind(addr string) error
		Close()
	}

	connector interface {
		context.Context
		GetEndpoint() Endpoint
		Connect(Endpoint)
	}

	listener interface {
		context.Context
		Listen() <-chan Endpoint
	}
)

type bindCtx struct{ context.Context }

func newBindCtx(addr string, p portal) *bindCtx {
	c := context.WithValue(p.c, keyBindAddr, addr)
	c = context.WithValue(c, keySvrEndpt, p)
	c = context.WithValue(c, keyListenChan, make(chan Endpoint))
	return &bindCtx{Context: c}
}

func (bc bindCtx) Addr() string            { return bc.Value(keyBindAddr).(string) }
func (bc bindCtx) l() chan Endpoint        { return bc.Value(keyListenChan).(chan Endpoint) }
func (bc bindCtx) GetEndpoint() Endpoint   { return bc.Value(keySvrEndpt).(Endpoint) }
func (bc bindCtx) Connect(ep Endpoint)     { bc.l() <- ep }
func (bc bindCtx) Listen() <-chan Endpoint { return bc.l() }
func (bc bindCtx) Close()                  { close(bc.l()) }

type trans struct {
	sync.RWMutex
	lookup map[string]*bindCtx
}

func (t *trans) GetConnector(a string) (c connector, ok bool) {
	t.RLock()
	c, ok = t.lookup[a]
	t.RUnlock()
	return
}

func (t *trans) GetListener(c context.Context) (listener, error) {
	t.Lock()
	defer t.Unlock()

	bc := &bindCtx{Context: c}

	if _, exists := t.lookup[bc.Addr()]; exists {
		return nil, errors.Errorf("transport exists at %s", bc.Addr())
	}

	t.lookup[bc.Addr()] = bc
	ctx.Defer(bc, func() {
		bc.Close()
		t.Lock()
		delete(t.lookup, bc.Addr())
		t.Unlock()
	})

	return bc, nil
}
