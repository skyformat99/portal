package benchpubsub

import (
	"testing"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/proto/pub"
	"github.com/lthibault/portal/proto/sub"
)

var (
	recvVal interface{}
	p0      = pub.New(portal.Cfg{})
	p1      = sub.New(portal.Cfg{})
)

func init() {
	var err error
	if err = p1.Bind("/"); err != nil {
		panic(err)
	}

	p1.Subscribe(sub.TopicAll)

	if err = p0.Connect("/"); err != nil {
		panic(err)
	}
}

func bench(i int, b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		go p0.Send(n)
		recvVal = p1.Recv()
	}
}

func Benchmark10(b *testing.B)         { bench(10, b) }
func Benchmark100(b *testing.B)        { bench(100, b) }
func Benchmark1000(b *testing.B)       { bench(1000, b) }
func Benchmark10000(b *testing.B)      { bench(10000, b) }
func Benchmark100000(b *testing.B)     { bench(100000, b) }
func Benchmark1000000(b *testing.B)    { bench(1000000, b) }
func Benchmark1000000000(b *testing.B) { bench(1000000000, b) }
