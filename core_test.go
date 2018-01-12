package portal

import (
	"testing"
	"time"

	"github.com/SentimensRG/ctx"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var ( // test type constraints
	_ Portal = &portal{}
)

type mockProto struct {
	mockProtoSig
	epAdded   chan Endpoint
	epRemoved chan Endpoint
}

func (m mockProto) Init(ProtocolPortal)        {}
func (m mockProto) AddEndpoint(ep Endpoint)    { m.epAdded <- ep }
func (m mockProto) RemoveEndpoint(ep Endpoint) { m.epRemoved <- ep }

func TestMkPortal(t *testing.T) {
	proto := mockProto{}

	t.Run("DonerAssignment", func(t *testing.T) {
		d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))
		defer cancel()

		cfg := Cfg{Doner: d}

		ptl := newPortal(proto, cfg, cancel)

		if ptl.Doner != d {
			t.Error("Doner instance passed to Cfg not assigned to portal")
		}
	})

	t.Run("Sync", func(t *testing.T) {
		d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))
		defer cancel()

		cfg := Cfg{Doner: d}

		ptl := newPortal(proto, cfg, cancel)

		if ptl.Size != 0 {
			t.Errorf("configuration error: size should default to 0, got %d", ptl.Size)
		}

		if ptl.Async() {
			t.Error("configuration error: portal is asynchronous")
		}
	})

	t.Run("Async", func(t *testing.T) {
		d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))
		defer cancel()

		cfg := Cfg{
			Doner: d,
			Size:  1,
		}

		ptl := newPortal(proto, cfg, cancel)

		if ptl.Size != 1 {
			t.Errorf("configuration error: expected size=1, got %d", ptl.Size)
		}

		if !ptl.Async() {
			t.Error("configuration error: portal is synchronous")
		}
	})
}

func TestTransportIntegration(t *testing.T) {
	var boundEP Endpoint

	t.Run("Bind", func(t *testing.T) {
		d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))

		cfg := Cfg{Doner: d}

		ptl := newPortal(mockProto{}, cfg, cancel)
		boundEP = ptl // will be tested in "Connect" test

		if err := ptl.Bind("/test"); err != nil {
			t.Errorf("failed to bind portal to address /test: %s", err)
		}

		if transport.lookup["/test"].GetEndpoint().(*portal) != ptl {
			t.Error("mismatch between bound endpoint and portal")
		}
	})

	t.Run("Connect", func(t *testing.T) {
		epAdded := make(chan Endpoint)
		epRemoved := make(chan Endpoint)

		proto := mockProto{
			epAdded:   epAdded,
			epRemoved: epRemoved,
		}

		d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))

		cfg := Cfg{Doner: d}

		ptl := newPortal(proto, cfg, cancel)

		// Connect calls portal.AddEndpoint, so we need to recv in order not to
		// block

		var g errgroup.Group
		var ep Endpoint
		g.Go(func() (err error) {
			ep = <-epAdded
			return
		})

		if err := ptl.Connect("/test"); err != nil {
			t.Errorf("failed to connect portal to address /test: %s", err)
		}

		if g.Wait(); ep != boundEP {
			t.Error("endpoint passed to protocol does not belong to bound portal")
		}
	})

	t.Run("EndpointTransaction", func(t *testing.T) {
		// BINDING PORTAL
		bindEPAdded := make(chan Endpoint)
		bindEPRemoved := make(chan Endpoint)

		bindProto := mockProto{
			epAdded:   bindEPAdded,
			epRemoved: bindEPRemoved,
		}

		dBind, dCancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))

		cfg := Cfg{Doner: dBind}

		bindP := newPortal(bindProto, cfg, dCancel)

		// CONNECTING PORTAL
		connEPAdded := make(chan Endpoint)
		connEPRemoved := make(chan Endpoint)

		connProto := mockProto{
			epAdded:   connEPAdded,
			epRemoved: connEPRemoved,
		}

		dConn, cCancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))

		cfg = Cfg{Doner: dConn}

		connP := newPortal(connProto, cfg, cCancel)

		// BIND
		if err := bindP.Bind("/XYZ"); err != nil {
			t.Errorf("failed to bind to addr /XYZ: %s", err)
		}

		t.Run("EndpointExchange", func(t *testing.T) {
			var g errgroup.Group

			g.Go(func() (err error) {
				select {
				case <-connEPAdded:
				case <-time.After(time.Millisecond * 100):
					err = errors.New("connEP timeout")
				}
				return
			})

			g.Go(func() (err error) {
				select {
				case <-bindEPAdded:
				case <-time.After(time.Millisecond * 100):
					err = errors.New("bindEP timeout")
				}
				return
			})

			// CONNECT:  THIS IS WHEN ENDPOINTS SHOULD BE *ADDED*
			if err := connP.Connect("/XYZ"); err != nil {
				t.Errorf("failed to connect to addr /XYZ: %s", err)
			}

			// TEST IF ENDPOINTS WERE RECEIVED
			if err := g.Wait(); err != nil {
				t.Error(err)
			}
		})

		t.Run("EndpointGC", func(t *testing.T) {
			var g errgroup.Group

			g.Go(func() (err error) {
				select {
				case <-connEPRemoved:
				case <-time.After(time.Millisecond * 100):
					err = errors.New("connEP timeout")
				}
				return
			})

			g.Go(func() (err error) {
				select {
				case <-bindEPRemoved:
				case <-time.After(time.Millisecond * 100):
					err = errors.New("bindEP timeout")
				}
				return
			})

			go bindP.Close()
			go connP.Close()

			// TEST IF ENDPOINTS WERE RECEIVED
			if err := g.Wait(); err != nil {
				t.Error(err)
			}
		})
	})
}

type mockProtoExt struct {
	mockProto
	onSend, onRecv func(*Message) bool
}

func (m mockProtoExt) SendHook(msg *Message) bool { return m.onSend(msg) }
func (m mockProtoExt) RecvHook(msg *Message) bool { return m.onRecv(msg) }

func mkSendRecvTestPortal(p Protocol, size int) (*portal, func()) {
	d, cancel := ctx.WithCancel(ctx.Lift(make(chan struct{})))
	cfg := Cfg{Doner: d, Size: size}
	ptl := newPortal(p, cfg, cancel)
	return ptl, cancel
}

func TestSendRecvClose(t *testing.T) {
	p := mockProtoExt{
		onSend: func(*Message) bool { return true },
		onRecv: func(*Message) bool { return true },
	}
	pDrop := mockProtoExt{
		onSend: func(*Message) bool { return false },
		onRecv: func(*Message) bool { return false },
	}

	t.Run("SendMsg", func(t *testing.T) {
		t.Run("Drop", func(t *testing.T) {
			ptl, _ := mkSendRecvTestPortal(pDrop, 1)
			defer ptl.Close()

			if err := ptl.Bind("/alpha"); err != nil {
				t.Errorf("failed to bind: %s", err)
			}

			msg := NewMsg()
			defer msg.wait()

			ptl.SendMsg(msg)
			select {
			case <-ptl.chSend:
				t.Error("value SHOULD NOT have been sent")
				msg.Free()
			default:
				if msg.refcnt != 0 {
					t.Error("message was dropped without being freed")
				}
			}
		})

		t.Run("DontDrop", func(t *testing.T) {
			ptl, _ := mkSendRecvTestPortal(p, 1)
			defer ptl.Close()

			if err := ptl.Bind("/bravo"); err != nil {
				t.Errorf("failed to bind: %s", err)
			}

			msg := NewMsg()
			defer msg.wait()
			msg.Value = true

			ptl.SendMsg(msg)
			select {
			case m := <-ptl.chSend:
				if m.refcnt != 1 {
					t.Errorf("message unexpectedly reffed or freed (refcnt=%d)", msg.refcnt)
				}
				if m.Value != msg.Value {
					t.Errorf("unexpected value in message (got %v, expected %v)", m.Value, msg.Value)
				}
				msg.Free()
			default:
				t.Error("value SHOULD have been sent")
			}
		})
	})

	t.Run("RecvMsg", func(t *testing.T) {
		t.Run("Drop", func(t *testing.T) {
			ptl, _ := mkSendRecvTestPortal(pDrop, 1)
			defer ptl.Close()

			if err := ptl.Bind("/charlie"); err != nil {
				t.Errorf("failed to bind: %s", err)
			}

			msg := NewMsg()
			defer msg.wait()
			ptl.chRecv <- msg

			ch := make(chan *Message)
			go func() { ch <- ptl.RecvMsg() }()

			select {
			case <-ch:
				t.Error("message SHOULD NOT have been sent")
				msg.Free()
			case <-time.After(time.Microsecond * 1):
				if msg.refcnt != 0 {
					t.Errorf("msg was dropped without being freed (refcnt=%d)", msg.refcnt)
				}
			}
		})

		t.Run("DontDrop", func(t *testing.T) {
			ptl, _ := mkSendRecvTestPortal(p, 1)
			defer ptl.Close()

			if err := ptl.Bind("/bravo"); err != nil {
				t.Errorf("failed to bind: %s", err)
			}

			msg := NewMsg()
			defer msg.wait()
			msg.Value = true
			ptl.chRecv <- msg

			ch := make(chan *Message)
			go func() { ch <- ptl.RecvMsg() }()

			select {
			case m := <-ch:
				if msg.refcnt != 1 {
					t.Errorf("unexpectedly reffed or freed message (expected 1, refcnt=%d)", msg.refcnt)
				}
				if m.Value.(bool) != msg.Value.(bool) {
					t.Errorf("unexpected value in message (expected %v, got %v)", msg.Value, m.Value)
				}
				msg.Free()
			case <-time.After(time.Microsecond * 1):
				t.Error("message SHOULD have been sent")
			}
		})
	})

	t.Run("Send", func(t *testing.T) {
	})

	t.Run("Recv", func(t *testing.T) {
	})

	t.Run("Close", func(t *testing.T) {
		ptl, _ := mkSendRecvTestPortal(mockProto{}, 1)
		ptl.Close()

		select {
		case <-ptl.Done():
		default:
			t.Error("call to close did not fire protocol's Doner")
		}
	})

}
