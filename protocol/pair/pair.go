package pair

import (
	"log"
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
		go p.startReceiving(p.peer)
		go p.startSending(p.peer)
	}
}

func (p *pair) RemoveEndpoint(ep portal.Endpoint) {
	p.Lock()
	defer p.Unlock()

	if peer := p.peer; peer != nil && peer.ep == ep {
		p.peer = nil
		close(peer.chHalt)
	}
}

func (*pair) Number() uint16     { return portal.ProtoPair }
func (*pair) Name() string       { return "pair" }
func (*pair) PeerNumber() uint16 { return portal.ProtoPair }
func (*pair) PeerName() string   { return "pair" }

func (p *pair) startReceiving(ep *pairEP) {
	for {
		msg := ep.ep.RecvMsg()
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

func (p *pair) startSending(ep *pairEP) {
	var msg *portal.Message
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ DEBUG ] %v", r)
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

			ep.ep.SendMsg(msg) // may panic

		case <-ep.chHalt:
			return
		case <-p.prtl.CloseChannel():
			return
		}
	}
}

// New allocates a Portal using the PAIR protocol
func New() portal.Portal {
	return portal.MakePortal(&pair{})
}
