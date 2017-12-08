package bus

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/protocol"
)

type bus struct {
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

func (b *bus) Init(ptl portal.ProtocolPortal) {
	b.ptl = ptl
	b.n = proto.NewNeighborhood()
	go b.startSending()
}

func (b bus) startSending() {
	sq := b.ptl.SendChannel()
	cq := b.ptl.CloseChannel()

	var wg sync.WaitGroup
	var msg *portal.Message
	for {
		select {
		case <-cq:
			return
		case msg = <-sq:
			if msg == nil {
				sq = b.ptl.SendChannel()
				continue
			}

			// broadcast
			m, done := b.n.RMap()
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

func (b bus) startReceiving(pe proto.PeerEndpoint) {
	var msg *portal.Message
	defer func() {
		if msg != nil {
			msg.Free()
		}
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	rq := b.ptl.RecvChannel()
	cq := b.ptl.CloseChannel()

	for msg = pe.Announce(); msg != nil; pe.Announce() {
		select {
		case rq <- msg:
		case <-cq:
			return
		}
	}
}

func (b bus) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(b, ep.Signature())
	pe := proto.NewPeerEP(ep)

	b.n.SetPeer(ep.ID(), pe)
}

func (b bus) RemoveEndpoint(ep portal.Endpoint) { b.n.DropPeer(ep.ID()) }

func (bus) Number() uint16     { return proto.Bus }
func (bus) PeerNumber() uint16 { return proto.Bus }
func (bus) Name() string       { return "bus" }
func (bus) PeerName() string   { return "bus" }

// New allocates a portal using the BUS protocol
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg, &bus{})
}
