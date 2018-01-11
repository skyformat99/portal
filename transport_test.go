package portal

import (
	"testing"
	"time"

	"github.com/SentimensRG/ctx"
	"github.com/satori/go.uuid"
)

type mockProtoSig struct {
	name, peerName     string
	number, peerNumber uint16
}

func (m mockProtoSig) Name() string       { return m.name }
func (m mockProtoSig) PeerName() string   { return m.peerName }
func (m mockProtoSig) Number() uint16     { return m.number }
func (m mockProtoSig) PeerNumber() uint16 { return m.peerNumber }

type mockEP struct {
	uuid.UUID
	sc  chan *Message
	rc  chan *Message
	sig mockProtoSig
}

func (m mockEP) ID() uuid.UUID                { return m.UUID }
func (m mockEP) Close()                       {}
func (m mockEP) RecvChannel() chan<- *Message { return m.rc }
func (m mockEP) SendChannel() <-chan *Message { return m.sc }
func (m mockEP) Signature() ProtocolSignature { return m.sig }

func TestTransport(t *testing.T) {
	// Overwrite global transport variable to facilitate tear-down & isolate
	// tests
	transport := trans{lookup: make(map[string]*binding)}

	epBind := &mockEP{
		UUID: uuid.Must(uuid.NewV4()),
		sig:  mockProtoSig{},
	}

	epConn := &mockEP{
		UUID: uuid.Must(uuid.NewV4()),
		sig:  mockProtoSig{},
	}

	d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))

	var l listener
	var c connector

	t.Run("GetListener", func(t *testing.T) {
		t.Run("BindSuccess", func(t *testing.T) {
			var err error
			if l, err = transport.GetListener(d, "/test", epBind); err != nil {
				t.Errorf("failed to bind endpoint to transport: %s", err)
			}

			if b, ok := transport.lookup["/test"]; !ok {
				t.Error("*binder not found at bound address /test")
			} else if id := b.GetEndpoint().ID(); id != epBind.UUID {
				t.Errorf("unexpected endpoint in transport lookup table (expected %s, got %s)", id, epBind.UUID)
			}
		})

		t.Run("BindFailure", func(t *testing.T) {
			if _, err := transport.GetListener(d, "/test", epBind); err == nil {
				t.Error("address-space collision not detected")
			}
		})
	})

	t.Run("GetConnector", func(t *testing.T) {
		t.Run("ConnectSuccess", func(t *testing.T) {
			var ok bool
			if c, ok = transport.GetConnector("/test"); !ok {
				t.Error("failed to retrieve connector from bound addess /test")
			}
		})

		t.Run("ConnectFailure", func(t *testing.T) {
			if _, ok := transport.GetConnector("/fail"); ok {
				t.Error("retrieval of non-existant connector reported as successful")
			}
		})

		t.Run("ConnectorListenerCorrespondence", func(t *testing.T) {
			if c.GetEndpoint() != epBind {
				t.Error("pointer mismatch between stored & retrieved endpoints")
			}
		})

	})

	t.Run("DoListenAndConnect", func(t *testing.T) {
		go c.Connect(epConn)
		if ep := <-l.Listen(); ep != epConn {
			t.Error("received unexpected endpoint")
		}
	})

	t.Run("GarbageCollection", func(t *testing.T) {
		b := transport.lookup["/test"]

		cancel()

		t.Run("BindingClosed", func(t *testing.T) {
			if _, ok := <-b.Listen(); ok {
				t.Error("listen channel not closed")
			}
		})

		t.Run("ReleaseBinding", func(t *testing.T) {
			select {
			case <-b.Done():
			case <-time.After(time.Millisecond * 200):
				t.Error("*binding did not release")
			}
		})

		t.Run("ReleaseConnector", func(t *testing.T) {
			select {
			case <-c.Done():
			case <-time.After(time.Millisecond * 200):
				t.Error("connector did not release")
			}
		})

		t.Run("RemoveLookupEntry", func(t *testing.T) {
			<-time.After(time.Millisecond * 1)
			if b, ok := transport.lookup["/test"]; ok {
				t.Errorf("*binding not removed (addr=%s, id=%s)", b.Addr(), b.GetEndpoint().ID())
			}
		})
	})
}

// t.Run("", func(t *testing.T) {
// })