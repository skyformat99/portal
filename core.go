package portal

import (
	"context"

	"github.com/SentimensRG/ctx/sigctx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// Cfg is a base configuration struct
type Cfg struct {
	Ctx context.Context
}

// MakePortal is for protocol implementations
func MakePortal(c context.Context, p Protocol) Portal {
	var cancel context.CancelFunc
	if c == nil {
		c = sigctx.New()
	}

	c, cancel = context.WithCancel(c)
	return newPortal(c, cancel, p)
}

type portal struct {
	id    uuid.UUID
	proto Protocol
	ready bool

	c      context.Context
	cancel context.CancelFunc

	chSend chan *Message
	chRecv chan *Message
}

func newPortal(c context.Context, cancel context.CancelFunc, p Protocol) (prtl *portal) {
	prtl = &portal{
		id:     uuid.NewV4(),
		proto:  p,
		c:      c,
		cancel: cancel,
		chSend: make(chan *Message),
		chRecv: make(chan *Message),
	}

	p.Init(prtl)
	return
}

func (p *portal) Connect(addr string) (err error) {
	c, ok := transport.GetConnector(addr)
	if !ok {
		return errors.New("connection refused")
	}

	c.Connect(p)
	go p.gc(c.Done(), c.GetEndpoint())

	p.ready = true
	go func() {
		<-p.c.Done()
		p.ready = false
	}()

	return
}

func (p *portal) Bind(addr string) (err error) {
	var l listener
	if l, err = transport.GetListener(newBindCtx(addr, *p)); err != nil {
		err = errors.Wrap(err, "portal bind error")
	} else {
		go func() {
			for ep := range l.Listen() {
				go p.gc(l.Done(), ep)
			}
		}()
	}

	p.ready = true
	go func() {
		<-p.c.Done()
		p.ready = false
	}()

	return
}

func (p portal) Send(v interface{}) {
	if !p.ready {
		panic(errors.New("send to disconnected portal"))
	}

	msg := NewMsg()
	msg.Value = v
	p.SendMsg(msg)
}

func (p portal) Recv() (v interface{}) {
	if !p.ready {
		panic(errors.New("recv from disconnected portal"))
	}

	if msg := p.RecvMsg(); msg != nil {
		v = msg.Value
		msg.Free()
	}

	return
}

func (p portal) Close() { p.cancel() }
func (p portal) SendMsg(msg *Message) {
	select {
	case p.chSend <- msg:
	case <-p.c.Done():
	}

}
func (p portal) RecvMsg() (msg *Message) {
	select {
	case msg = <-p.chRecv:
	case <-p.c.Done():
	}
	return
}

// Implement Endpoint
func (p portal) ID() uuid.UUID                { return p.id }
func (p portal) Notify(msg *Message)          { p.chRecv <- msg }
func (p portal) Announce() *Message           { return <-p.chSend }
func (p portal) Signature() ProtocolSignature { return p.proto }

// Implement ProtocolSocket
func (p portal) SendChannel() <-chan *Message  { return p.chSend }
func (p portal) RecvChannel() chan<- *Message  { return p.chRecv }
func (p portal) CloseChannel() <-chan struct{} { return p.c.Done() }

// gc manages the lifecycle of an endpoint
func (p portal) gc(chRemoteDone <-chan struct{}, ep Endpoint) {
	p.proto.AddEndpoint(ep)

	select {
	case <-p.c.Done():
	case <-chRemoteDone:
	}

	p.proto.RemoveEndpoint(ep)
}
