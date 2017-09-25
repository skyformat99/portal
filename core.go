package portal

import (
	"github.com/pkg/errors"
)

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

	c, ok := router.GetConnector(addr)
	if !ok {
		return errors.New("connection refused")
	}

	c.Connect(p)
	go p.gc(c.GetEndpoint())
	return
}

func (p portal) Bind(addr string) (err error) {
	var l listener
	if l, err = router.GetListener(addr, p); err != nil {
		err = errors.Wrap(err, "portal bind error")
	} else {
		go func() {
			<-p.chHalt
			l.Close()
		}()

		go func() {
			for ep := range l.Listen() {
				go p.gc(ep)
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

// gc manages the lifecycle of an endpoint
func (p portal) gc(ep Endpoint) {
	p.proto.AddEndpoint(ep)
	<-p.chHalt
	p.proto.RemoveEndpoint(ep)
}
