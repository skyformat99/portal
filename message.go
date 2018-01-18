package portal

import (
	"sync"
)

const (
	// HeaderSenderID identifies the sender's ID
	HeaderSenderID HdrKey = iota
)

var (
	msgPool = messagePool{
		Pool: sync.Pool{New: func() interface{} { return newMsg().Ref() }},
	}
)

type messagePool struct{ sync.Pool }

func (pool *messagePool) Get() *Message    { return pool.Pool.Get().(*Message) }
func (pool *messagePool) Put(msg *Message) { go pool.put(msg) }
func (pool *messagePool) put(msg *Message) {
	for k := range msg.Header {
		delete(msg.Header, k)
	}
	pool.Pool.Put(msg)
}

// HdrKey is a key for a Header
type HdrKey uint8

// Header of message
type Header map[HdrKey]interface{}

// GetHeader the value associated with a key
func (h Header) GetHeader(key HdrKey) (v interface{}, ok bool) {
	v, ok = h[key]
	return
}

// SetHeader a value to a key
func (h Header) SetHeader(key HdrKey, v interface{}) { h[key] = v }

// Message wraps a value and sends it down the portal
type Message struct {
	wg sync.WaitGroup
	Header
	Value interface{}
}

func newMsg() *Message { return &Message{Header: Header(make(map[HdrKey]interface{}, 1))} }

// Free deallocates a message
func (m *Message) Free() { m.wg.Done() }

// Ref increments the reference count on the message.  Note that since the
// underlying message is actually shared, consumers must take care not
// to modify the message.  Applications should *NOT* make use of this
// function -- it is intended for Protocol, Transport and internal use only.
func (m *Message) Ref() *Message {
	m.wg.Add(1)
	return m
}

// Wait for the message to be delivered.  It MUST be called by the Send function.
func (m *Message) wait() {
	m.wg.Wait()
	msgPool.Put(m.Ref())
}

// NewMsg returns a message with a single refcount
func NewMsg() *Message { return msgPool.Get() }
