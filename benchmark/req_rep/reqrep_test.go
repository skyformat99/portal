package benchreqrep

import (
	"testing"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/proto/rep"
	"github.com/lthibault/portal/proto/req"
)

var (
	i       uint64
	recvVal uint64
	p0      = req.New(portal.Cfg{})
	p1      = rep.New(portal.Cfg{})
)

func init() {
	var err error
	if err = p0.Bind("/"); err != nil {
		panic(err)
	}

	if err = p1.Connect("/"); err != nil {
		panic(err)
	}
}

func increment() {
	i = p1.Recv().(uint64)
	i++
	p1.Send(i)
}

func send() {
	go p0.Send(recvVal)
	recvVal = p0.Recv().(uint64)
}

func bench(i int, b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		go send()
		increment()
	}
}

func Benchmark10(b *testing.B)         { bench(10, b) }
func Benchmark100(b *testing.B)        { bench(100, b) }
func Benchmark1000(b *testing.B)       { bench(1000, b) }
func Benchmark10000(b *testing.B)      { bench(10000, b) }
func Benchmark100000(b *testing.B)     { bench(100000, b) }
func Benchmark1000000(b *testing.B)    { bench(1000000, b) }
func Benchmark1000000000(b *testing.B) { bench(1000000000, b) }
