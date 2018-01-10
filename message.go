package portal

import (
	"sync"
	"sync/atomic"
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
	if v := atomic.AddInt32(&m.refcnt, -1); v == 0 {
		m.cond.Signal()
		atomic.StoreInt32(&m.refcnt, 1)
		msgPool.Put(m)
	} else if v < 0 {
		panic("unreachable")
	}
}

// Ref increments the reference count on the message.  Note that since the
// underlying message is actually shared, consumers must take care not
// to modify the message.  Applications should *NOT* make use of this
// function -- it is intended for Protocol, Transport and internal use only.
func (m *Message) Ref() { atomic.AddInt32(&m.refcnt, 1) }

// wait for the message to be delivered
func (m *Message) Wait() {
	m.cond.L.Lock()
	m.cond.Wait()
	m.cond.L.Unlock()
}

// NewMsg returns a message with a single refcount
func NewMsg() *Message { return msgPool.Get() }
