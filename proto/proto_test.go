package proto

import (
	"testing"
	"time"

	"github.com/satori/go.uuid"
)

// A little copying is better than a little dependency ...
type mockProtoSig struct {
	name, peerName     string
	number, peerNumber uint16
}

func (m mockProtoSig) Name() string       { return m.name }
func (m mockProtoSig) PeerName() string   { return m.peerName }
func (m mockProtoSig) Number() uint16     { return m.number }
func (m mockProtoSig) PeerNumber() uint16 { return m.peerNumber }

func TestEndpointsCompatible(t *testing.T) {
	if !EndpointsCompatible(mockProtoSig{}, mockProtoSig{}) {
		t.Error("protocols erroneously reported as incompatible")
	}

	if !EndpointsCompatible(mockProtoSig{number: 1}, mockProtoSig{peerNumber: 1}) {
		t.Error("protocols erroneously reported as incompatible")
	}
}

func TestPeerEndpoint(t *testing.T) {
	pe := NewPeerEP(nil)
	pe.Close()

	select {
	case <-pe.Done():
	default:
		t.Error("call to Close did not release the Done channel")
	}
}

func TestNeighborhood(t *testing.T) {
	n := NewNeighborhood().(*neighborhood)
	u := uuid.Must(uuid.NewV4())

	t.Run("SetPeer", func(t *testing.T) {
		n.SetPeer(u, nil)
		if _, ok := n.epts[u]; !ok {
			t.Error("PeerEndpoint not stored")
		}
	})

	t.Run("GetPeer", func(t *testing.T) {
		if _, ok := n.GetPeer(u); !ok {
			t.Error("failed to retrieve exiting peer")
		}
	})

	t.Run("DropPeer", func(t *testing.T) {
		n.DropPeer(u)
		if _, ok := n.epts[u]; ok {
			t.Error("drop operation did not evict PeerEndpoint from map")
		}
	})

	t.Run("RMap", func(t *testing.T) {
		for i := 0; i < 8; i++ {
			n.SetPeer(uuid.Must(uuid.NewV4()), nil)
		}

		var i int
		m, unlock := n.RMap()
		for _ = range m {
			i++
		}
		unlock()

		ch := make(chan struct{})
		go func() {
			n.Lock()
			n.Unlock()
			close(ch)
		}()

		select {
		case <-ch:
		case <-time.After(time.Microsecond * 1):
			t.Error("could not cycle lock; is `unlock` properly releasing it?")
		}

		if i != 8 {
			t.Errorf("expected 8 values, got %d", i)
		}
	})
}
