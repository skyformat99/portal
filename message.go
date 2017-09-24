package portal

import (
	"sync"
	"sync/atomic"
)

// TODO : rename to something like "packet" or "payload"

var msgPool = sync.Pool{New: func() interface{} {
	return &Message{
		Header: make(map[string]interface{}),
		refcnt: 1,
	}
}}

// Message wraps a value and sends it down the portal
type Message struct {
	Header map[string]interface{}
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
func NewMsg() *Message { return msgPool.Get().(*Message) }
