package portal

import (
	"sync"
	"sync/atomic"
)

var msgPool = messagePool{
	Pool: sync.Pool{New: func() interface{} {
		return &Message{refcnt: 1, delivered: make(chan struct{})}
	}},
}

type messagePool struct{ sync.Pool }

func (pool *messagePool) Get() *Message    { return pool.Pool.Get().(*Message) }
func (pool *messagePool) Put(msg *Message) { pool.Pool.Put(msg) }

// Message wraps a value and sends it down the portal
type Message struct {
	Value     interface{}
	refcnt    int32
	delivered chan struct{}
}

// Free deallocates a message
func (m *Message) Free() {
	if v := atomic.AddInt32(&m.refcnt, -1); v <= 0 {
		select {
		case m.delivered <- struct{}{}:
		default:
		}

		msgPool.Put(m)
	}
}

// Ref increments the reference count on the message.  Note that since the
// underlying message is actually shared, consumers must take care not
// to modify the message.  Applications should *NOT* make use of this
// function -- it is intended for Protocol, Transport and internal use only.
func (m *Message) Ref() { atomic.AddInt32(&m.refcnt, 1) }

// Signal that a message was delivered
func (m *Message) Signal(c chan<- struct{}) { c <- (<-m.delivered) }

// NewMsg returns a message with a single refcount
func NewMsg() *Message { return msgPool.Get() }
