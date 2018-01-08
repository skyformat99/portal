package pub

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/proto"
)

// Protocol implementing PUB
type Protocol struct {
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

// Init the Protocol
func (p *Protocol) Init(ptl portal.ProtocolPortal) {
	p.ptl = ptl
	p.n = proto.NewNeighborhood()
	go p.startSending()
}

func (p Protocol) startSending() {
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

func (p Protocol) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())

	pe := proto.NewPeerEP(ep)
	p.n.SetPeer(ep.ID(), pe)
}

func (p Protocol) RemoveEndpoint(ep portal.Endpoint) { p.n.DropPeer(ep.ID()) }

func (Protocol) Number() uint16     { return proto.Pub }
func (Protocol) PeerNumber() uint16 { return proto.Sub }
func (Protocol) Name() string       { return "pub" }
func (Protocol) PeerName() string   { return "sub" }

// New allocates a portal using the PUB protocol
func New(cfg portal.Cfg) portal.WriteOnly {
	return struct{ portal.WriteOnly }{portal.MakePortal(cfg, &Protocol{})} // write guard
}
