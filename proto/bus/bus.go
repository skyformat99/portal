package bus

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/proto"
)

// Protocol implementing BUS
type Protocol struct {
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

// Init the protocol (called by portal)
func (p *Protocol) Init(ptl portal.ProtocolPortal) {
	p.ptl = ptl
	p.n = proto.NewNeighborhood()
	go p.startSending()
}

func (p Protocol) startSending() {
	sq := p.ptl.SendChannel()
	cq := p.ptl.CloseChannel()

	var wg sync.WaitGroup // TODO:  optimize with CAS
	var msg *portal.Message
	for {
		select {
		case <-cq:
			return
		case msg = <-sq:
			if msg == nil {
				sq = p.ptl.SendChannel()
				continue
			}

			// broadcast
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

func (p Protocol) startReceiving(pe proto.PeerEndpoint) {
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

	for msg = pe.Announce(); msg != nil; pe.Announce() {
		select {
		case rq <- msg:
		case <-cq:
			return
		}
	}
}

func (p Protocol) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())
	pe := proto.NewPeerEP(ep)

	p.n.SetPeer(ep.ID(), pe)
}

func (p Protocol) RemoveEndpoint(ep portal.Endpoint) { p.n.DropPeer(ep.ID()) }

func (Protocol) Number() uint16     { return proto.Bus }
func (Protocol) PeerNumber() uint16 { return proto.Bus }
func (Protocol) Name() string       { return "bus" }
func (Protocol) PeerName() string   { return "bus" }

// New allocates a portal using the BUS protocol
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg, &Protocol{})
}
