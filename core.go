package portal

import (
	"github.com/SentimensRG/ctx"
	"github.com/SentimensRG/ctx/sigctx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// Cfg is a base configuration struct
type Cfg struct {
	ctx.Doner
	Size int
}

// Async returns true if the Portal is buffered
func (c Cfg) Async() bool { return c.Size > 0 }

// MakePortal is for protocol implementations
func MakePortal(cfg Cfg, p Protocol) Portal {
	var cancel func()
	var d ctx.Doner
	if cfg.Doner == nil {
		d = sigctx.New()
	}

	cfg.Doner, cancel = ctx.WithCancel(d)
	return newPortal(p, cfg, cancel)
}

type portal struct {
	Cfg
	cancel func()

	id    uuid.UUID
	proto Protocol
	ready bool

	txn    chan struct{}
	chSend chan *Message
	chRecv chan *Message
}

func newPortal(p Protocol, cfg Cfg, cancel func()) *portal {
	c := make(chan struct{})
	if cfg.Async() {
		close(c)
	}

	var ptl = new(portal)

	ptl.Cfg = cfg
	ptl.cancel = cancel
	ptl.id = uuid.Must(uuid.NewV4())
	ptl.txn = c
	ptl.proto = p
	ptl.chSend = make(chan *Message, cfg.Size)
	ptl.chRecv = make(chan *Message, cfg.Size)

	p.Init(ptl)

	return ptl
}

func (p *portal) Connect(addr string) (err error) {
	c, ok := transport.GetConnector(addr)
	if !ok {
		return errors.New("connection refused")
	}

	c.Connect(p)
	p.gc(c, c.GetEndpoint())

	p.ready = true
	go func() {
		<-p.Done()
		p.ready = false
	}()

	return
}

func (p *portal) Bind(addr string) (err error) {
	var l listener
	if l, err = transport.GetListener(newBinding(p, addr, p)); err != nil {
		err = errors.Wrap(err, "portal bind error")
	} else {
		go func() {
			for ep := range l.Listen() {
				p.gc(l, ep)
			}
		}()
	}

	p.ready = true
	ctx.Defer(p, func() { p.ready = false })

	return
}

func (p *portal) Send(v interface{}) {
	if !p.ready {
		panic(errors.New("send to disconnected portal"))
	}

	msg := NewMsg()
	msg.Value = v
	if !p.Async() {
		go msg.Signal(p.txn)
	}

	p.SendMsg(msg)
	<-p.txn // will be closed if portal is async
}

func (p *portal) Recv() (v interface{}) {
	if !p.ready {
		panic(errors.New("recv from disconnected portal"))
	}

	if msg := p.RecvMsg(); msg != nil {
		v = msg.Value
		msg.Free()
	}

	return
}

func (p *portal) Close() { p.cancel() }
func (p *portal) SendMsg(msg *Message) {
	select {
	case p.chSend <- msg:
	case <-p.Done():
	}
}

func (p *portal) RecvMsg() (msg *Message) {
	select {
	case msg = <-p.chRecv:
	case <-p.Done():
	}
	return
}

// Implement Endpoint
func (p *portal) ID() uuid.UUID                { return p.id }
func (p *portal) Notify(msg *Message)          { p.chRecv <- msg }
func (p *portal) Announce() *Message           { return <-p.chSend }
func (p *portal) Signature() ProtocolSignature { return p.proto }

// Implement ProtocolSocket
func (p *portal) SendChannel() <-chan *Message  { return p.chSend }
func (p *portal) RecvChannel() chan<- *Message  { return p.chRecv }
func (p *portal) CloseChannel() <-chan struct{} { return p.Done() }

// gc manages the lifecycle of an endpoint in the background
func (p *portal) gc(remote ctx.Doner, ep Endpoint) {
	p.proto.AddEndpoint(ep)
	ctx.Defer(ctx.Lift(ctx.Link(p, remote)), func() {
		p.proto.RemoveEndpoint(ep)
	})
}
