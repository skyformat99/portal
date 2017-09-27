package rep

import (
	"sync"

	"github.com/lthibault/portal"
	proto "github.com/lthibault/portal/protocol/core"
)

const eqBufSize = 8

type rep struct {
	sync.Mutex
	prtl portal.ProtocolPortal
	n    proto.Neighborhood
}

func (r *rep) Init(prtl portal.ProtocolPortal) {
	r.prtl = prtl
	r.n = proto.NewNeighborhood()
}

func (r *rep) startReceiving(pe proto.PeerEndpoint) {
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

	cq := r.prtl.CloseChannel()
	rq := r.prtl.RecvChannel()
	sq := r.prtl.SendChannel()

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
				pe.Notify(msg)
				r.Unlock()
			}
		}
	}
}

func (*rep) Number() uint16     { return portal.ProtoRep }
func (*rep) PeerNumber() uint16 { return portal.ProtoReq }
func (*rep) Name() string       { return "rep" }
func (*rep) PeerName() string   { return "req" }

func (r *rep) AddEndpoint(ep portal.Endpoint) {
	portal.MustBeCompatible(r, ep.Signature())

	pe := proto.NewPeerEP(ep)
	r.n.SetPeer(ep.ID(), pe)

	go r.startReceiving(pe)
}

func (r *rep) RemoveEndpoint(ep portal.Endpoint) { r.n.DropPeer(ep.ID()) }

// New allocates a new REP portal
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg.Ctx, &rep{})
}
