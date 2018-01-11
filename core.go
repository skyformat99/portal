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

	chSend chan *Message
	chRecv chan *Message

	ProtocolSendHook
	ProtocolRecvHook
}

func newPortal(p Protocol, cfg Cfg, cancel func()) *portal {
	var ptl = new(portal)

	ptl.Cfg = cfg
	ptl.cancel = cancel
	ptl.id = uuid.Must(uuid.NewV4())
	ptl.proto = p
	ptl.chSend = make(chan *Message, cfg.Size)
	ptl.chRecv = make(chan *Message, cfg.Size)

	if i, ok := interface{}(p).(ProtocolSendHook); ok {
		ptl.ProtocolSendHook = i.(ProtocolSendHook)
	}
	if i, ok := interface{}(p).(ProtocolRecvHook); ok {
		ptl.ProtocolRecvHook = i.(ProtocolRecvHook)
	}

	p.Init(ptl)

	return ptl
}

func (p *portal) Connect(addr string) (err error) {
	c, ok := transport.GetConnector(addr)
	if !ok {
		return errors.New("connection refused")
	}

	c.Connect(p)
	p.trackEndpoint(c, c.GetEndpoint())

	p.ready = true
	ctx.Defer(p, func() { p.ready = false })

	return
}

func (p *portal) Bind(addr string) (err error) {
	var l listener
	if l, err = transport.GetListener(p, addr, p); err != nil {
		err = errors.Wrap(err, "portal bind error")
	} else {
		go func() {
			for ep := range l.Listen() {
				p.trackEndpoint(l, ep)
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

	p.SendMsg(msg)

	if p.Async() {
		go msg.wait()
	} else {
		msg.wait()
	}
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
	if (p.ProtocolSendHook != nil) && !p.SendHook(msg) {
		msg.Free()
		return // drop msg silently
	}

	select {
	case p.chSend <- msg:
	case <-p.Done():
	}
}

func (p *portal) RecvMsg() (msg *Message) {
	for {
		select {
		case msg = <-p.chRecv:
			if (p.ProtocolRecvHook != nil) && !p.SendHook(msg) {
				msg.Free()
			} else {
				return
			}
		case <-p.Done():
			return
		}
	}
}

// Implement Endpoint
func (p *portal) ID() uuid.UUID { return p.id }

// func (p *portal) Notify(msg *Message)          { p.chRecv <- msg }
// func (p *portal) Announce() *Message           { return <-p.chSend }
func (p *portal) Signature() ProtocolSignature { return p.proto }

// Implement ProtocolSocket
func (p *portal) SendChannel() <-chan *Message  { return p.chSend }
func (p *portal) RecvChannel() chan<- *Message  { return p.chRecv }
func (p *portal) CloseChannel() <-chan struct{} { return p.Done() }

// gc manages the lifecycle of an endpoint in the background
func (p *portal) trackEndpoint(remote ctx.Doner, ep Endpoint) {
	p.proto.AddEndpoint(ep)
	ctx.Defer(ctx.Link(p, remote), func() { p.proto.RemoveEndpoint(ep) })
}
