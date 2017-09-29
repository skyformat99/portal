package test

import (
	"fmt"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol/pair"
)

func ExampleSendRecv() {
	p0 := pair.New(portal.Cfg{})
	p1 := pair.New(portal.Cfg{})
	_ = p0.Bind("/test")
	_ = p1.Connect("/test")
	defer p0.Close()
	defer p1.Close()

	go func() {
		s := p1.Recv().(string)
		p1.Send(s + ", world!")
	}()

	p0.Send("hello")
	fmt.Println(p0.Recv().(string))
	// Output: hello, world!
}
