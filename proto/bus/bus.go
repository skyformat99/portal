package bus

import (
	"sync"

	"github.com/SentimensRG/ctx"
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
	var wg sync.WaitGroup // TODO:  optimize with CAS (?)

	sq := p.ptl.SendChannel()
	cq := p.ptl.CloseChannel()

	for {
		select {
		case <-cq:
			return
		case msg, ok := <-sq:
			if !ok {
				// This should never happen.  If it does, the channels were not
				// closed in the correct order
				// TODO:  remove once tested & stable
				panic("ensure portal.Doner fires closes before chSend/chRecv")
			}

			// broadcast
			m, done := p.n.RMap()
			wg.Add(len(m))

			for _, peer := range m {
				go unicast(&wg, peer, msg)
			}

			done()
			wg.Wait()
		}
	}
}

func unicast(wg *sync.WaitGroup, pe proto.PeerEndpoint, msg *portal.Message) {
	select {
	case pe.RecvChannel() <- msg.Ref():
	case <-pe.Done():
	}
	wg.Done()
}

func (p Protocol) startReceiving(pe proto.PeerEndpoint) {
	rq := p.ptl.RecvChannel()
	cq := ctx.Link(ctx.Lift(p.ptl.CloseChannel()), pe)
	for msg := range pe.SendChannel() {
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
	go p.startReceiving(pe)
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
