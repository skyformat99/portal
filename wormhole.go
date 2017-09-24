package portal

import (
	"sync"

	"github.com/pkg/errors"
)

var wormhole = router{lookup: make(map[string]chan Endpoint)}

type router struct {
	sync.RWMutex
	lookup map[string]chan Endpoint
}

func (r *router) Get(s string) (ep chan<- Endpoint, ok bool) {
	r.RLock()
	defer r.RUnlock()

	ep, ok = r.lookup[s]
	return
}

func (r *router) Open(s string) (<-chan Endpoint, error) {
	r.Lock()
	defer r.Unlock()

	if _, exists := r.lookup[s]; exists {
		return nil, errors.Errorf("wormhole exists at %s", s)
	}

	ch := make(chan Endpoint)
	r.lookup[s] = ch
	return ch, nil
}

func (r *router) Collapse(s string) (ok bool) {
	r.Lock()
	defer r.Unlock()

	var wh chan<- Endpoint
	if wh, ok = r.lookup[s]; ok {
		close(wh)
	}

	return
}
