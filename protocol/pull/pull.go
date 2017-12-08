package pull

import (
	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol"
)

type pull struct{ ptl portal.ProtocolPortal }

func (p *pull) Init(ptl portal.ProtocolPortal) { p.ptl = ptl }

func (p pull) startReceiving(ep portal.Endpoint) {
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

	for msg = ep.Announce(); msg != nil; ep.Announce() {
		select {
		case rq <- msg:
		case <-cq:
			return
		}
	}
}

func (pull) Number() uint16                    { return proto.Pull }
func (pull) Name() string                      { return "pull" }
func (pull) PeerNumber() uint16                { return proto.Push }
func (pull) PeerName() string                  { return "push" }
func (pull) RemoveEndpoint(ep portal.Endpoint) {}

func (p pull) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())
	go p.startReceiving(ep)
}

// New allocates a Portal using the PULL protocol
func New(cfg portal.Cfg) portal.ReadOnly {
	return struct{ portal.ReadOnly }{portal.MakePortal(cfg, &pull{})} // read guard

}
