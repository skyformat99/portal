package portal

import (
	"sync"
	"sync/atomic"
)

// TODO : rename to something like "packet" or "payload"

const defaultHeaderSize = 8

// HeaderKey identifies a field in message headers
type HeaderKey uint32

var msgPool = messagePool{
	Pool: sync.Pool{New: func() interface{} {
		return &Message{
			refcnt: 1,
			Header: make(map[HeaderKey]interface{}, defaultHeaderSize),
		}
	}},
}

type messagePool struct{ sync.Pool }

func (pool *messagePool) Get() *Message { return pool.Pool.Get().(*Message) }

func (pool *messagePool) Put(msg *Message) {
	for k := range msg.Header {
		delete(msg.Header, k)
	}

	pool.Pool.Put(msg)
}

// Message wraps a value and sends it down the portal
type Message struct {
	Header map[HeaderKey]interface{}
	Value  interface{}

	refcnt int32
}

// Free deallocates a message
func (m *Message) Free() {
	if v := atomic.AddInt32(&m.refcnt, -1); v <= 0 {
		msgPool.Put(m)
	}
}

// NewMsg returns a message with a single refcount
func NewMsg() *Message { return msgPool.Get() }
