package pub

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/protocol/core"
)

type pub struct {
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

func (p *pub) Init(ptl portal.ProtocolPortal) {
	p.ptl = ptl
	p.n = proto.NewNeighborhood()
	go p.startSending()
}

func (p pub) startSending() {
	cq := p.ptl.CloseChannel()
	sq := p.ptl.SendChannel()

	var wg sync.WaitGroup

	var msg *portal.Message
	for {
		select {
		case <-cq:
			return
		case msg = <-sq:
			if msg == nil {
				sq = p.ptl.SendChannel()
			} else {
				m, done := p.n.RMap()
				wg.Add(len(m))

				for _, peer := range m {
					go func(peer proto.PeerEndpoint) {
						msg.Ref()
						peer.Notify(msg)
						wg.Done()
					}(peer)
				}

				done()
				msg.Free()
				wg.Wait()
			}
		}
	}
}

func (p pub) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())

	pe := proto.NewPeerEP(ep)
	p.n.SetPeer(ep.ID(), pe)
}

func (p pub) RemoveEndpoint(ep portal.Endpoint) { p.n.DropPeer(ep.ID()) }

func (pub) Number() uint16     { return proto.Pub }
func (pub) PeerNumber() uint16 { return proto.Sub }
func (pub) Name() string       { return "pub" }
func (pub) PeerName() string   { return "sub" }

// New allocates a portal using the PUB protocol
func New(cfg portal.Cfg) portal.WriteOnly {
	return proto.WriteGuard(portal.MakePortal(cfg, &pub{}))
}
