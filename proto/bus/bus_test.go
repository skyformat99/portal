package bus

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/lthibault/portal"
)

func TestIntegration(t *testing.T) {
	const nPtls = 4

	ptls := make([]portal.Portal, nPtls)
	for i := range ptls {
		ptls[i] = New(portal.Cfg{})
	}

	bP, cP := ptls[0], ptls[1:len(ptls)]

	if err := bP.Bind("/test/bus/integration"); err != nil {
		t.Error(err)
	}

	for i, p := range cP {
		if err := p.Connect("/test/bus/integration"); err != nil {
			t.Errorf("portal %d: %s", i, err)
		}
	}

	var wg sync.WaitGroup

	ch := make(chan struct{})
	go func() {
		for {
			wg.Add(nPtls - 1)
			wg.Wait()
			ch <- struct{}{}
		}
	}()

	go func() {
		for {
			log.Printf("bindPortal recved %b", bP.Recv())
			wg.Done()
		}
	}()

	for i, p := range cP {
		go func(i int, p portal.Portal) {
			for {
				log.Printf("connPortal %d recved %b", i, p.Recv())
				wg.Done()
			}
		}(i, p)
	}

	t.Run("SendBind", func(t *testing.T) {
		go bP.Send(true)
		select {
		case <-ch:
		case <-time.After(time.Millisecond * 100):
			t.Error("at least one portal did not receive the value sent by bP")
		}
	})

	for i, p := range cP {
		t.Run(fmt.Sprintf("SendPortal%d", i), func(t *testing.T) {
			go p.Send(true)
			select {
			case <-ch:
			case <-time.After(time.Millisecond * 100):
				t.Error("at least one portal did not receive the value sent by bP")
			}
		})
	}
}
