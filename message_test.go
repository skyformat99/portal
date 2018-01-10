package portal

import (
	"testing"
	"time"
)

func TestMessagePool(t *testing.T) {
	m := NewMsg()

	t.Run("TestRefIncr", func(t *testing.T) {
		if m.refcnt != 1 {
			t.Errorf("refcnt not initialized to 1 (%d)", m.refcnt)
		}

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
