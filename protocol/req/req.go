package req

import (
	"sync"

	"github.com/lthibault/portal"
	"github.com/satori/go.uuid"
)

type req struct {
	sync.Mutex

	prtl portal.ProtocolPortal
	epts map[uuid.UUID]*reqEP
}

type reqEP struct {
	ep     portal.Endpoint
	chHalt chan struct{}
}

func (r *req) Init(prtl portal.ProtocolPortal) {
	r.prtl = prtl
	r.epts = make(map[uuid.UUID]*reqEP, 8)
}

func (r *req) startSending(pe *reqEP) {

	// NB: Because this function is only called when an endpoint is
	// added, we can reasonably safely cache the channels -- they won't
	// be changing after this point.

	sq := r.prtl.SendChannel()
	cq := r.prtl.CloseChannel()

	var msg *portal.Message
	for {
		select {
		case msg = <-sq:
		case <-cq:
			return
		case <-pe.chHalt:
			return
		}

		pe.ep.Notify(msg)
	}
}

func (r *req) startReceiving(ep portal.Endpoint) {
	rq := r.prtl.RecvChannel()
	cq := r.prtl.CloseChannel()

	var msg *portal.Message
	for {
		if msg = ep.Announce(); msg == nil {
			break
		}

		select {
		case rq <- msg:
		case <-cq:
			msg.Free()
			break
		}
	}
}

func (*req) Number() uint16     { return portal.ProtoReq }
func (*req) PeerNumber() uint16 { return portal.ProtoRep }
func (*req) Name() string       { return "req" }
func (*req) PeerName() string   { return "rep" }

func (r *req) AddEndpoint(ep portal.Endpoint) {
	portal.MustBeCompatible(r, ep.Signature())

	pe := &reqEP{ep: ep, chHalt: make(chan struct{})}

	r.Lock()
	r.epts[ep.ID()] = pe
	r.Unlock()

	go r.startSending(pe)
	go r.startReceiving(ep)
}

func (r *req) RemoveEndpoint(ep portal.Endpoint) {
	id := ep.ID()

	r.Lock()
	pe := r.epts[id]
	r.Unlock()

	if pe != nil {
		close(pe.chHalt)
	}
}

// New allocates a Portal using the REQ protocol
func New(cfg portal.Cfg) portal.Portal {
	return portal.MakePortal(cfg.Ctx, &req{})
}
