package portal

import (
	"context"
	"sync"

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
	ctx := context.WithValue(p.ctx, keyBindAddr, addr)
	ctx = context.WithValue(ctx, keySvrEndpt, p)
	ctx = context.WithValue(ctx, keyListenChan, make(chan Endpoint))
	return &bindCtx{Context: ctx}
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
	defer t.RUnlock()

	c, ok = t.lookup[a]
	return
}

func (t *trans) GetListener(ctx context.Context) (listener, error) {
	t.Lock()
	defer t.Unlock()

	bc := &bindCtx{Context: ctx}

	if _, exists := t.lookup[bc.Addr()]; exists {
		return nil, errors.Errorf("transport exists at %s", bc.Addr())
	}

	t.lookup[bc.Addr()] = bc
	go func() {
		<-bc.Done()
		bc.Close()
		t.Lock()
		delete(t.lookup, bc.Addr())
		t.Unlock()
	}()

	return bc, nil
}
