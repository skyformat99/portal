package portal

import "github.com/pkg/errors"

// MakePortal is for protocol implementations
func MakePortal(p Protocol) Portal {
	return newPortal(p)
}

type portal struct {
	proto  Protocol
	chHalt chan struct{}
	chSend chan *Message
	chRecv chan *Message
}

func newPortal(p Protocol) (prtl *portal) {
	prtl = &portal{
		proto:  p,
		chHalt: make(chan struct{}),
		chSend: make(chan *Message),
		chRecv: make(chan *Message),
	}

	p.Init(prtl)
	return
}

func (p portal) Connect(addr string) (err error) {
	wHole, ok := wormhole.Get(addr)
	if !ok {
		err = errors.New("connection refused")
	} else {
		wHole <- p
		// BUG : DEADLOCK
		// the server never receives an endpoint from our side
	}

	return
}

func (p portal) Bind(addr string) (err error) {
	// This function sets up goroutines to accept inbound connections.
	// The accepted connection will be added to a list of accepted connections.
	// The socket just needs to listen continuously as we assume that we want
	// to continue to receive inbound connections without limit.

	var wHole <-chan Endpoint
	if wHole, err = wormhole.Open(addr); err != nil {
		err = errors.Wrapf(err, "could not open wormhole")
	} else {
		go func() {
			<-p.chHalt
			wormhole.Collapse(addr)
		}()

		go func() {
			for ep := range wHole {
				p.proto.AddEndpoint(ep)
			}
		}()
	}

	return
}

func (p portal) Send(v interface{}) {
	msg := NewMsg()
	msg.Value = v
	p.SendMsg(msg)
}

func (p portal) Recv() (v interface{}) {
	msg := p.RecvMsg()
	v = msg.Value
	msg.Free()
	return
}

func (p portal) Close()               { close(p.chHalt) }
func (p portal) SendMsg(msg *Message) { p.chSend <- msg }
func (p portal) RecvMsg() *Message    { return <-p.chRecv }

// Implement ProtocolSocket
func (p portal) SendChannel() <-chan *Message  { return p.chSend }
func (p portal) RecvChannel() chan<- *Message  { return p.chRecv }
func (p portal) CloseChannel() <-chan struct{} { return p.chHalt }
