package pair

import (
	"sync"

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

func (p *pair) AddEndpoint(ep portal.Endpoint) {
	portal.MustBeCompatible(p, ep.Signature())

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
		ep.Close()
		p.peer = nil
		close(peer.chHalt)
	}
}

func (*pair) Number() uint16     { return portal.ProtoPair }
func (*pair) Name() string       { return "pair" }
func (*pair) PeerNumber() uint16 { return portal.ProtoPair }
func (*pair) PeerName() string   { return "pair" }

func (p *pair) startReceiving() {
	var msg *portal.Message
	rq := p.prtl.RecvChannel()
	cq := p.prtl.CloseChannel()

	for {
		if msg = p.peer.ep.Announce(); msg == nil {
			return // upstream channel was closed
		}

		select {
		case rq <- msg:
		case <-cq:
			return
		}
	}
}

func (p *pair) startSending() {
	var msg *portal.Message
	defer func() {
		if r := recover(); r != nil {
			msg.Free()
			panic(r)
		}
	}()

	sq := p.prtl.SendChannel()
	cq := p.prtl.CloseChannel()

	// This is pretty easy because we have only one peer at a time.
	// If the peer goes away, drop the message on the floor.
	for {
		select {
		case <-p.peer.chHalt:
			return
		case <-cq:
			return
		case msg = <-sq:
			if msg == nil {
				sq = p.prtl.SendChannel()
			} else {
				p.peer.ep.Notify(msg) // may panic
			}
		}
	}
}

// New allocates a Portal using the PAIR protocol
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg.Ctx, &pair{})
}
