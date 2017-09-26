package rep

import (
	"sync"

	"github.com/satori/go.uuid"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol"
)

const defaultEptsSize = 8

type repEP struct {
	chHalt chan struct{}
	ep     portal.Endpoint
}

type rep struct {
	sync.Mutex
	prtl portal.ProtocolPortal
	epts map[uuid.UUID]*repEP
}

func (r *rep) Init(prtl portal.ProtocolPortal) {
	r.prtl = prtl
	r.epts = make(map[uuid.UUID]*repEP, defaultEptsSize)
	go r.startSending()
}

func (r *rep) startSending() {
	cq := r.prtl.CloseChannel()
	sq := r.prtl.SendChannel()

	var msg *portal.Message
	for {
		select {
		case msg = <-sq:
			if msg == nil {
				sq = r.prtl.SendChannel()
				continue
			}
		case <-cq:
			return
		}

		r.Lock()
		re := r.epts[msg.Header[protocol.REQEndpt].(uuid.UUID)]
		r.Unlock()

		if re == nil {
			msg.Free()
		} else {
			re.ep.Notify(msg)
		}
	}
}

func (r *rep) startReceiving(ep portal.Endpoint) {
	var msg *portal.Message
	defer func() {
		if r := recover(); r != nil {
			// caller might want to catch panics.  avoid memleak
			msg.Free()
			panic(r)
		}
	}()

	rq := r.prtl.RecvChannel()
	cq := r.prtl.CloseChannel()

	for {
		if msg = ep.Announce(); msg == nil {
			return
		}

		msg.Header[protocol.REQEndpt] = ep.ID()

		select {
		case rq <- msg:
		case <-cq:
			msg.Free()
			return
		}
	}
}

func (*rep) Number() uint16     { return portal.ProtoRep }
func (*rep) PeerNumber() uint16 { return portal.ProtoReq }
func (*rep) Name() string       { return "rep" }
func (*rep) PeerName() string   { return "req" }

func (r *rep) AddEndpoint(ep portal.Endpoint) {
	portal.MustBeCompatible(r, ep.Signature())

	re := &repEP{ep: ep, chHalt: make(chan struct{})}

	r.Lock()
	r.epts[ep.ID()] = re
	r.Unlock()

	go r.startReceiving(ep)
	go r.startSending()
}

func (r *rep) RemoveEndpoint(ep portal.Endpoint) {
	id := ep.ID()

	r.Lock()
	re := r.epts[id]
	delete(r.epts, id)
	r.Unlock()

	if re != nil {
		close(re.chHalt)
	}
}

// New allocates a new REP portal
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg.Ctx, &rep{})
}
