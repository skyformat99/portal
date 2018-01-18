package push

import (
	"github.com/SentimensRG/ctx"
	"github.com/lthibault/portal"
	"github.com/lthibault/portal/proto"
)

// Protocol implementing PUSH
type Protocol struct {
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

// Init the PUSH protocol
func (p *Protocol) Init(ptl portal.ProtocolPortal) {
	close(ptl.RecvChannel()) // NOTE : if mysterious error, maybe it's this?
	p.ptl = ptl
	p.n = proto.NewNeighborhood()
}

func (p Protocol) startSending(pe proto.PeerEndpoint) {
	sq := p.ptl.SendChannel()
	rq := pe.RecvChannel()
	cq := ctx.Link(ctx.Lift(p.ptl.CloseChannel()), pe)

	for {
		select {
		case <-cq:
			return
		case msg, ok := <-sq:
			if !ok {
				sq = p.ptl.SendChannel()
			} else {
				rq <- msg
			}
		}
	}
}

func (Protocol) Number() uint16     { return proto.Push }
func (Protocol) Name() string       { return "push" }
func (Protocol) PeerNumber() uint16 { return proto.Pull }
func (Protocol) PeerName() string   { return "pull" }

func (p Protocol) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())

	pe := proto.NewPeerEP(ep)
	p.n.SetPeer(ep.ID(), pe)
	go p.startSending(pe)
}

func (p Protocol) RemoveEndpoint(ep portal.Endpoint) { p.n.DropPeer(ep.ID()) }

// New allocates a WriteOnly Portal using the PUSH protocol
func New(cfg portal.Cfg) portal.WriteOnly {
	return struct{ portal.WriteOnly }{portal.MakePortal(cfg, &Protocol{})} // write guard
}
