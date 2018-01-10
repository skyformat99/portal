package portal

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

var (
	msgPool = messagePool{
		Pool: sync.Pool{New: func() interface{} {
			return &Message{
				cond:   &sync.Cond{L: &sync.Mutex{}},
				refcnt: 1,
			}
		}},
	}
)

type messagePool struct{ sync.Pool }

func (pool *messagePool) Get() *Message    { return pool.Pool.Get().(*Message) }
func (pool *messagePool) Put(msg *Message) { pool.Pool.Put(msg) }

// Message wraps a value and sends it down the portal
type Message struct {
	cond   *sync.Cond
	Value  interface{}
	refcnt int32
}

// Free deallocates a message
func (m *Message) Free() {
	if atomic.AddInt32(&m.refcnt, -1) < 0 {
		panic(errors.Errorf("unreachable: ref count < 0 (%d)", m.refcnt))
	}
}

// Ref increments the reference count on the message.  Note that since the
// underlying message is actually shared, consumers must take care not
// to modify the message.  Applications should *NOT* make use of this
// function -- it is intended for Protocol, Transport and internal use only.
func (m *Message) Ref() { atomic.AddInt32(&m.refcnt, 1) }

// Wait for the message to be delivered.  It MUST be called by the Send function.
func (m *Message) wait() {
	for atomic.CompareAndSwapInt32(&m.refcnt, 0, 1) { // block
	}

	msgPool.Put(m)
}

// NewMsg returns a message with a single refcount
func NewMsg() *Message { return msgPool.Get() }
