package portal

import (
	"testing"
	"time"

	"github.com/SentimensRG/ctx"
)

func TestTransport(t *testing.T) {
	// Overwrite global transport variable to facilitate tear-down & isolate
	// tests
	transport := trans{lookup: make(map[string]*binding)}

	epBind := &mockEP{
		id:  NewID(),
		sig: mockProtoSig{},
	}

	epConn := &mockEP{
		id:  NewID(),
		sig: mockProtoSig{},
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
			} else if id := b.GetEndpoint().ID(); id != epBind.id {
				t.Errorf("unexpected endpoint in transport lookup table (expected %s, got %s)", id, epBind.id)
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
