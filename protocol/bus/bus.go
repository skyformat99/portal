package bus

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/protocol/core"
)

type bus struct {
	prtl portal.ProtocolPortal
	n    proto.Neighborhood
}

func (b *bus) Init(prtl portal.ProtocolPortal) {
	b.prtl = prtl
	b.n = proto.NewNeighborhood()
	go b.startSending()
}

func (b bus) startSending() {
	sq := b.prtl.SendChannel()
	cq := b.prtl.CloseChannel()

	var wg sync.WaitGroup
	var msg *portal.Message
	for {
		select {
		case <-cq:
			return
		case msg = <-sq:
			if msg == nil {
				sq = b.prtl.SendChannel()
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

	rq := b.prtl.RecvChannel()
	cq := b.prtl.CloseChannel()

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
	return portal.MakePortal(cfg.Ctx, &bus{})
}
