package push

import (
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/proto/pull"
)

const iter = 10000

func TestIntegration(t *testing.T) {
	const nPtls = 4

	pushP := New(portal.Cfg{})
	if err := pushP.Bind("/test/push/integration"); err != nil {
		t.Error(err)
	}

	ptls := make([]portal.ReadOnly, nPtls)
	for i := range ptls {
		ptls[i] = pull.New(portal.Cfg{})
		if err := ptls[i].Connect("/test/push/integration"); err != nil {
			t.Error(err)
		}
	}

	var g errgroup.Group

	go func() {
		for i := 0; i < iter; i++ {
			pushP.Send(i)
		}
	}()

	var zero, one, two, three bool

	fn := func(n int, p portal.ReadOnly) func() error {
		return func() (err error) {
			for i := 0; i < iter/nPtls; i++ {
				switch _ = p.Recv().(int); n {
				case 0:
					zero = true
				case 1:
					one = true
				case 2:
					two = true
				case 3:
					three = true
				}
			}
			return
		}
	}

	for i, p := range ptls {
		g.Go(fn(i, p))
	}

	ch := make(chan struct{})
	go func() {
		g.Wait()
		close(ch)
	}()

	select {
	case <-ch:
		if !(zero && one && two && three) {
			t.Error("at least one pull-portal did not recv a value")
		}
	case <-time.After(time.Millisecond * 100):
		t.Error("at least one pull-portal blocked (N.B.:  this error is often stochastic, appearing after recompilations)")
	}

}
