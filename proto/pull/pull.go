package pull

import (
	"github.com/lthibault/portal"
	"github.com/lthibault/portal/proto"
)

// Protocol implementing PULL
type Protocol struct{ ptl portal.ProtocolPortal }

func (p *Protocol) Init(ptl portal.ProtocolPortal) { p.ptl = ptl }

func (p Protocol) startReceiving(ep portal.Endpoint) {
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

func (Protocol) Number() uint16                    { return proto.Pull }
func (Protocol) Name() string                      { return "pull" }
func (Protocol) PeerNumber() uint16                { return proto.Push }
func (Protocol) PeerName() string                  { return "push" }
func (Protocol) RemoveEndpoint(ep portal.Endpoint) {}

func (p Protocol) AddEndpoint(ep portal.Endpoint) {
	proto.MustBeCompatible(p, ep.Signature())
	go p.startReceiving(ep)
}

// New allocates a Portal using the PULL protocol
func New(cfg portal.Cfg) portal.ReadOnly {
	return struct{ portal.ReadOnly }{portal.MakePortal(cfg, &Protocol{})} // read guard

}
