package portal

import (
	"testing"
	"time"
)

func TestMessagePool(t *testing.T) {
	m := NewMsg()

	t.Run("TestRefInit", func(t *testing.T) {
		if m.refcnt != 1 {
			t.Errorf("refcnt not initialized to 1 (%d)", m.refcnt)
		}
	})

	t.Run("TestRefIncr", func(t *testing.T) {
		if m.Ref(); m.refcnt != 2 {
			t.Errorf("refcnt not incremented (%d)", m.refcnt)
		}
	})

	t.Run("TestRefDecr", func(t *testing.T) {
		if m.Free(); m.refcnt != 1 {
			t.Errorf("refcnt not decremented (%d)", m.refcnt)
		}
	})

	t.Run("TestWait", func(t *testing.T) {
		ch := make(chan struct{})
		go func() {
			m.wait()
			close(ch)
		}()

		select {
		case <-ch:
			t.Errorf("call to wait did not block (refcnt=%d)", m.refcnt)
		case <-time.After(time.Millisecond):
		}

		m.Free()

		select {
		case <-ch:
			if m.refcnt != 1 {
				t.Errorf("refcnt not reset to 1 (%d)", m.refcnt)
			}
		case <-time.After(time.Second * 1):
			t.Error("Message.wait() did not return")
		}
	})

}
