package push

import (
	"sync"

	"github.com/lthibault/portal"
	"github.com/satori/go.uuid"
)

type push struct {
	sync.Mutex
	prtl portal.ProtocolPortal
	epts map[uuid.UUID]*pushEP
}

type pushEP struct {
	ep     portal.Endpoint
	chHalt chan struct{}
}

func (p *push) Init(prtl portal.ProtocolPortal) {
	p.prtl = prtl
	p.epts = make(map[uuid.UUID]*pushEP, 1) // usually we only have 1 PULLer
}

func (p *push) startSending(ep *pushEP) {
	var msg *portal.Message
	defer func() {
		if r := recover(); r != nil {
			msg.Free()
		}
	}()

	sq := p.prtl.SendChannel()
	cq := p.prtl.CloseChannel()

	for {
		select {
		case <-cq:
			return
		case <-ep.chHalt:
			return
		case msg = <-sq:
			if msg == nil {
				sq = p.prtl.SendChannel()
			} else {
				ep.ep.Notify(msg)
			}
		}
	}
}

func (*push) Number() uint16     { return portal.ProtoPush }
func (*push) Name() string       { return "push" }
func (*push) PeerNumber() uint16 { return portal.ProtoPull }
func (*push) PeerName() string   { return "pull" }

func (p *push) AddEndpoint(ep portal.Endpoint) {
	close(p.prtl.RecvChannel()) // NOTE : if mysterious error, maybe it's this?
	pe := &pushEP{ep: ep, chHalt: make(chan struct{})}

	p.Lock()
	p.epts[ep.ID()] = pe
	p.Unlock()

	go p.startSending(pe)
}

func (p *push) RemoveEndpoint(ep portal.Endpoint) {
	id := ep.ID()

	p.Lock()
	pe := p.epts[id]
	delete(p.epts, id)
	p.Unlock()

	if pe != nil {
		close(pe.chHalt)
	}
}

// New allocates a WriteOnly Portal using the PUSH protocol
func New(cfg portal.Cfg) portal.WriteOnly {

	// The anonymous "guard" struct prevents users from recasting push portals
	// into read-write portals through type assertions or type switches.
	// For example, `myPushPortal.(portal.Portal)` will fail.

	return struct{ portal.WriteOnly }{
		WriteOnly: portal.MakePortal(cfg.Ctx, &push{}),
	}
}
