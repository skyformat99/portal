package pair

import (
	"testing"

	"github.com/lthibault/portal"
	"golang.org/x/sync/errgroup"
)

const iter = 10000

func TestIntegration(t *testing.T) {
	p0 := New(portal.Cfg{})
	p1 := New(portal.Cfg{})

	if err := p0.Bind("/test/pair/integration"); err != nil {
		t.Error(err)
	}

	if err := p1.Connect("/test/pair/integration"); err != nil {
		t.Error(err)
	}

	var g errgroup.Group

	var l2r, r2l int

	// Test Left to Right
	g.Go(func() (err error) {
		for i := 0; i < iter; i++ {
			p0.Send(i)
		}
		return
	})

	g.Go(func() (err error) {
		for i := 0; i < iter; i++ {
			l2r = p1.Recv().(int)
		}
		return
	})

	// Test Right to Left
	g.Go(func() (err error) {
		for i := 0; i < iter; i++ {
			p1.Send(i)
		}
		return
	})

	g.Go(func() (err error) {
		for i := 0; i < iter; i++ {
			r2l = p0.Recv().(int)
		}
		return
	})

	if err := g.Wait(); err != nil {
		t.Error(err)
	}

	if l2r != iter-1 {
		t.Errorf("left to right:  expected %d, got %d", iter-1, l2r)
	}

	if r2l != iter-1 {
		t.Errorf("right to left:  expected %d, got %d", iter-1, r2l)
	}
}
