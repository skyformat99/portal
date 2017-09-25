package portal

import (
	"context"

	"github.com/SentimensRG/sigctx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// Cfg is a base configuration struct
type Cfg struct {
	Ctx context.Context
}

// MakePortal is for protocol implementations
func MakePortal(ctx context.Context, p Protocol) Portal {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx = sigctx.New()
	}

	ctx, cancel = context.WithCancel(ctx)
	return newPortal(ctx, cancel, p)
}

type portal struct {
	id    uuid.UUID
	proto Protocol

	ctx    context.Context
	cancel context.CancelFunc

	chSend chan *Message
	chRecv chan *Message
}

func newPortal(ctx context.Context, c context.CancelFunc, p Protocol) (prtl *portal) {
	prtl = &portal{
		id:     uuid.NewV4(),
		proto:  p,
		ctx:    ctx,
		cancel: c,
		chSend: make(chan *Message),
		chRecv: make(chan *Message),
	}

	p.Init(prtl)
	return
}

func (p portal) Connect(addr string) (err error) {

	c, ok := transport.GetConnector(addr)
	if !ok {
		return errors.New("connection refused")
	}

	c.Connect(p)
	go p.gc(c.Done(), c.GetEndpoint())
	return
}

func (p portal) Bind(addr string) (err error) {
	ctx := context.WithValue(p.ctx, keyBindAddr, addr)
	ctx = context.WithValue(ctx, keySvrEndpt, p)
	ctx = context.WithValue(ctx, keyListenChan, make(chan Endpoint))

	var l listener
	if l, err = transport.GetListener(ctx); err != nil {
		err = errors.Wrap(err, "portal bind error")
	} else {
		go func() {
			for ep := range l.Listen() {
				go p.gc(l.Done(), ep)
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

func (p portal) Close()               { p.cancel() }
func (p portal) SendMsg(msg *Message) { p.chSend <- msg }
func (p portal) RecvMsg() *Message    { return <-p.chRecv }

// Implement Endpoint
func (p portal) ID() uuid.UUID       { return p.id }
func (p portal) Notify(msg *Message) { p.chRecv <- msg }
func (p portal) Announce() *Message  { return <-p.chSend }

// Implement ProtocolSocket
func (p portal) SendChannel() <-chan *Message  { return p.chSend }
func (p portal) RecvChannel() chan<- *Message  { return p.chRecv }
func (p portal) CloseChannel() <-chan struct{} { return p.ctx.Done() }

// gc manages the lifecycle of an endpoint
func (p portal) gc(chRemoteDone <-chan struct{}, ep Endpoint) {
	p.proto.AddEndpoint(ep)

	select {
	case <-p.ctx.Done():
	case <-chRemoteDone:
	}

	p.proto.RemoveEndpoint(ep)
}
