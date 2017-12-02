package portal

import (
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/pkg/errors"
)

var transport = trans{lookup: make(map[string]*binding)}

type (
	transporter interface {
		Connect(addr string) error
		Bind(addr string) error
		Close()
	}

	connector interface {
		ctx.Doner
		GetEndpoint() Endpoint
		Connect(Endpoint)
	}

	listener interface {
		ctx.Doner
		Listen() <-chan Endpoint
	}
)

type binding struct {
	ctx.Doner
	addr  string
	cxns  chan Endpoint // incomming connections
	bound Endpoint
}

func newBinding(addr string, p portal) *binding {
	return &binding{Doner: p.d, addr: addr, cxns: make(chan Endpoint), bound: p}
}

func (b binding) Addr() string            { return b.addr }
func (b binding) GetEndpoint() Endpoint   { return b.bound }
func (b binding) Connect(ep Endpoint)     { b.cxns <- ep }
func (b binding) Listen() <-chan Endpoint { return b.cxns }
func (b binding) Close()                  { close(b.cxns) }

type trans struct {
	sync.RWMutex
	lookup map[string]*binding
}

func (t *trans) GetConnector(a string) (c connector, ok bool) {
	t.RLock()
	c, ok = t.lookup[a]
	t.RUnlock()
	return
}

func (t *trans) GetListener(b *binding) (listener, error) {
	t.Lock()
	defer t.Unlock()

	if _, exists := t.lookup[b.Addr()]; exists {
		return nil, errors.Errorf("transport exists at %s", b.Addr())
	}

	t.lookup[b.Addr()] = b
	ctx.Defer(b, func() {
		b.Close()
		t.Lock()
		delete(t.lookup, b.Addr())
		t.Unlock()
	})

	return b, nil
}
