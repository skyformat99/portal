package proto

import (
	"sync"

	"github.com/SentimensRG/ctx"
	"github.com/lthibault/portal"
	"github.com/pkg/errors"
)

const (
	Pair = iota
	Push
	Pull
	Req
	Rep
	Pub
	Sub
	Surv
	Resp
	Bus
	Star
	Brok
	Deal
)

// EndpointsCompatible returns true if the Endpoints have compatible protocols
func EndpointsCompatible(sig0, sig1 portal.ProtocolSignature) bool {
	return sig0.Number() == sig1.PeerNumber() && sig0.Number() == sig1.PeerNumber()
}

// MustBeCompatible panics if the Endpoints have incompatible protocols
func MustBeCompatible(sig0, sig1 portal.ProtocolSignature) {
	if !EndpointsCompatible(sig0, sig1) {
		panic(errors.Errorf("%s incompatible with %s", sig0.Name(), sig1.Name()))
	}
}

// PeerEndpoint is the endpoint to a remote peer.
type PeerEndpoint interface {
	ctx.Doner
	portal.Endpoint
}

type peerEP struct {
	portal.Endpoint
	cq chan struct{}
}

// NewPeerEP wraps an endpoint in a PeerEndpoint
func NewPeerEP(ep portal.Endpoint) PeerEndpoint {
	return &peerEP{Endpoint: ep, cq: make(chan struct{})}
}

func (p peerEP) Done() <-chan struct{} { return p.cq }
func (p peerEP) Close()                { close(p.cq) }

// Neighborhood maintains a map of PeerEndpoints
type Neighborhood interface {
	RMap() (map[ID]PeerEndpoint, func())
	SetPeer(ID, PeerEndpoint)
	GetPeer(ID) (PeerEndpoint, bool)
	DropPeer(ID)
}

// neighborhood stores connected peer endpoints
type neighborhood struct {
	sync.RWMutex
	epts map[ID]PeerEndpoint
}

// NewNeighborhood initializes a Neighborhood
func NewNeighborhood() Neighborhood {
	return &neighborhood{epts: make(map[ID]PeerEndpoint)}
}

func (n *neighborhood) RMap() (map[ID]PeerEndpoint, func()) {
	n.RLock()
	return n.epts, n.RUnlock
}

func (n *neighborhood) SetPeer(id ID, pe PeerEndpoint) {
	n.Lock()
	n.epts[id] = pe
	n.Unlock()
}

func (n *neighborhood) GetPeer(id ID) (p PeerEndpoint, ok bool) {
	n.RLock()
	p, ok = n.epts[id]
	n.RUnlock()
	return
}

func (n *neighborhood) DropPeer(id ID) {
	n.Lock()
	pe := n.epts[id]
	delete(n.epts, id)
	n.Unlock()

	if pe != nil {
		pe.Close()
	}
}
