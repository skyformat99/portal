package push

import (
	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol/core"
)

type push struct {
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

func (p *push) Init(ptl portal.ProtocolPortal) {
	p.ptl = ptl
	p.n = proto.NewNeighborhood()
}

func (p push) startSending(pe proto.PeerEndpoint) {
	var msg *portal.Message
	defer func() {
		if r := recover(); r != nil {
			msg.Free()
			panic(r)
		}
	}()

	sq := p.ptl.SendChannel()
	cq := p.ptl.CloseChannel()

	for {
		select {
		case <-cq:
			return
		case <-pe.Done():
			return
		case msg = <-sq:
			if msg == nil {
				sq = p.ptl.SendChannel()
			} else {
				pe.Notify(msg)
			}
		}
	}
}

func (push) Number() uint16     { return proto.Push }
func (push) Name() string       { return "push" }
func (push) PeerNumber() uint16 { return proto.Pull }
func (push) PeerName() string   { return "pull" }

func (p push) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())
	close(p.ptl.RecvChannel()) // NOTE : if mysterious error, maybe it's this?

	pe := proto.NewPeerEP(ep)
	p.n.SetPeer(ep.ID(), pe)
	go p.startSending(pe)
}

func (p push) RemoveEndpoint(ep portal.Endpoint) { p.n.DropPeer(ep.ID()) }

// New allocates a WriteOnly Portal using the PUSH protocol
func New(cfg portal.Cfg) portal.WriteOnly {
	return proto.WriteGuard(portal.MakePortal(cfg, &push{}))
}
