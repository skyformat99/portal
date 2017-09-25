package pair

import (
	"sync"
	"time"

	"github.com/lthibault/portal"
)

type pair struct {
	sync.Mutex
	prtl portal.ProtocolPortal
	peer *pairEP
}

type pairEP struct {
	ep     portal.Endpoint
	chHalt chan struct{}
}

func (p *pair) Init(prtl portal.ProtocolPortal) { p.prtl = prtl }
func (p *pair) Shutdown(expire time.Time)       { panic("NOT IMPLEMENTED") }

func (p *pair) AddEndpoint(ep portal.Endpoint) {
	p.Lock()
	defer p.Unlock()

	if p.peer != nil { // we already have a conn, reject this one
		ep.Close()
	} else {
		p.peer = &pairEP{chHalt: make(chan struct{}), ep: ep}
		go p.startReceiving()
		go p.startSending()
	}
}

func (p *pair) RemoveEndpoint(ep portal.Endpoint) {
	p.Lock()
	defer p.Unlock()

	if peer := p.peer; peer != nil && peer.ep.ID() == ep.ID() {
		p.peer = nil
		close(peer.chHalt)
	}
}

func (*pair) Number() uint16     { return portal.ProtoPair }
func (*pair) Name() string       { return "pair" }
func (*pair) PeerNumber() uint16 { return portal.ProtoPair }
func (*pair) PeerName() string   { return "pair" }

func (p *pair) startReceiving() {
	for {
		msg := p.peer.ep.Announce()
		if msg == nil {
			return // upstream channel was closed
		}

		select {
		case p.prtl.RecvChannel() <- msg:
		case <-p.prtl.CloseChannel():
			return
		}
	}
}

func (p *pair) startSending() {
	var msg *portal.Message
	defer func() {
		if r := recover(); r != nil {
			msg.Free()
		}
	}()

	// This is pretty easy because we have only one peer at a time.
	// If the peer goes away, drop the message on the floor.
	for {
		select {
		case msg = <-p.prtl.SendChannel():
			if msg == nil {
				continue
			}

			p.peer.ep.Notify(msg) // may panic

		case <-p.peer.chHalt:
			return
		case <-p.prtl.CloseChannel():
			return
		}
	}
}

// New allocates a Portal using the PAIR protocol
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg.Ctx, &pair{})
}
