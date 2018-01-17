package portal

import (
	"testing"
	"time"
)

func TestMessagePool(t *testing.T) {
	m := NewMsg()
	ch := make(chan struct{})
	go func() {
		m.wait()
		close(ch)
	}()

	select {
	case <-ch:
		t.Error("call to wait did not block")
	case <-time.After(time.Millisecond):
	}

	m.Free()

	select {
	case <-ch:
	case <-time.After(time.Second * 1):
		t.Error("wait did not return")
	}

}
