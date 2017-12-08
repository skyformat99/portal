package pair

import (
	"sync"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol"
)

type pair struct {
	sync.Mutex
	ptl  portal.ProtocolPortal
	peer proto.PeerEndpoint
}

func (p *pair) Init(ptl portal.ProtocolPortal) { p.ptl = ptl }

func (p *pair) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())

	p.Lock()
	defer p.Unlock()

	if p.peer != nil { // we already have a conn, reject this one
		ep.Close()
	} else {
		p.peer = proto.NewPeerEP(ep)
		go p.startReceiving()
		go p.startSending()
	}
}

func (p *pair) RemoveEndpoint(ep portal.Endpoint) {
	p.Lock()
	defer p.Unlock()

	if peer := p.peer; peer != nil && peer.ID() == ep.ID() {
		ep.Close()
		p.peer = nil
		peer.Close()
	}
}

func (*pair) Number() uint16     { return proto.Pair }
func (*pair) Name() string       { return "pair" }
func (*pair) PeerNumber() uint16 { return proto.Pair }
func (*pair) PeerName() string   { return "pair" }

func (p *pair) startReceiving() {
	var msg *portal.Message
	defer func() {
		if msg != nil {
			msg.Free()
		}
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	rq := p.ptl.RecvChannel()
	cq := p.ptl.CloseChannel()

	for msg = p.peer.Announce(); msg != nil; p.peer.Announce() {
		select {
		case rq <- msg:
		case <-cq:
			return
		}
	}
}

func (p *pair) startSending() {
	sq := p.ptl.SendChannel()
	cq := p.ptl.CloseChannel()
	pcq := p.peer.Done()

	// This is pretty easy because we have only one peer at a time.
	// If the peer goes away, drop the message on the floor.
	var msg *portal.Message
	for {
		select {
		case <-pcq:
			return
		case <-cq:
			return
		case msg = <-sq:
			if msg == nil {
				sq = p.ptl.SendChannel()
			} else {
				p.peer.Notify(msg) // may panic
			}
		}
	}
}

// New allocates a Portal using the PAIR protocol
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg, &pair{})
}
