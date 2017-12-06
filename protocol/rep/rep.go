package rep

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/protocol/core"
)

const eqBufSize = 8

type rep struct {
	sync.Mutex
	ptl portal.ProtocolPortal
	n   proto.Neighborhood
}

func (r *rep) Init(ptl portal.ProtocolPortal) {
	r.ptl = ptl
	r.n = proto.NewNeighborhood()
}

func (r *rep) startServing(pe proto.PeerEndpoint) {
	var msg *portal.Message
	defer func() {
		if msg != nil {
			msg.Free()
		}
		if r := recover(); r != nil {
			msg.Free()
			panic(r)
		}
	}()

	cq := r.ptl.CloseChannel()
	rq := r.ptl.RecvChannel()
	sq := r.ptl.SendChannel()

	for msg = pe.Announce(); msg != nil; msg = pe.Announce() {
		r.Lock()

		select {
		case <-cq:
			r.Unlock()
			return
		case rq <- msg:
			select {
			case <-cq:
				r.Unlock()
				return
			case <-pe.Done():
				msg.Free()
				r.Unlock()
				continue
			case msg = <-sq:
				if msg == nil {
					sq = r.ptl.SendChannel()
					r.Unlock()
				} else {
					pe.Notify(msg)
					r.Unlock()
				}
			}
		}
	}
}

func (*rep) Number() uint16     { return proto.Rep }
func (*rep) PeerNumber() uint16 { return proto.Req }
func (*rep) Name() string       { return "rep" }
func (*rep) PeerName() string   { return "req" }

func (r *rep) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(r, ep.Signature())

	pe := proto.NewPeerEP(ep)
	r.n.SetPeer(ep.ID(), pe)

	go r.startServing(pe)
}

func (r *rep) RemoveEndpoint(ep portal.Endpoint) { r.n.DropPeer(ep.ID()) }

// New allocates a new REP portal
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg, &rep{})
}
