package star

import (
	"github.com/SentimensRG/ctx"
	"github.com/lthibault/portal"
)

const starAddrFmt = "star:///%s"

type rawPortal interface {
	ctx.Doner
	Bind(string) error
	Connect(string) error
	SendMsg(*portal.Message)
	RecvMsg() *portal.Message
	Close()
}

// Protocol implementing STAR
type Protocol struct {
	ptl     portal.ProtocolPortal
	bus     rawPortal
	busAddr string
}

// Init the protocol (called by portal)
func (p *Protocol) Init(ptl portal.ProtocolPortal) { p.ptl = ptl }
