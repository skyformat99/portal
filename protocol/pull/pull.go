package pull

import "github.com/lthibault/portal"

type pull struct{ prtl portal.ProtocolPortal }

func (p *pull) Init(prtl portal.ProtocolPortal) { p.prtl = prtl }

func (p pull) startReceiving(ep portal.Endpoint) {
	var msg *portal.Message

	rq := p.prtl.RecvChannel()
	cq := p.prtl.CloseChannel()

	for {
		if msg = ep.Announce(); msg == nil {
			return
		}

		select {
		case rq <- msg:
		case <-cq:
			return
		}
	}
}

func (*pull) Number() uint16                    { return portal.ProtoPull }
func (*pull) Name() string                      { return "pull" }
func (*pull) PeerNumber() uint16                { return portal.ProtoPush }
func (*pull) PeerName() string                  { return "push" }
func (*pull) RemoveEndpoint(ep portal.Endpoint) {}

func (p *pull) AddEndpoint(ep portal.Endpoint) {
	portal.MustBeCompatible(p, ep.Signature())
	go p.startReceiving(ep)
}

// New allocates a Portal using the PULL protocol
func New(cfg portal.Cfg) portal.ReadOnly {

	// The anonymous "guard" struct prevents users from recasting pull portals
	// into read-write portals through type assertions or type switches.
	// For example, `myPushPortal.(portal.Portal)` will fail.

	return struct{ portal.ReadOnly }{
		ReadOnly: portal.MakePortal(cfg.Ctx, &pull{}),
	}
}
