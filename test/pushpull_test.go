package testpp

import (
	"testing"

	"github.com/lthibault/portal"
	"github.com/lthibault/portal/protocol/pull"
	"github.com/lthibault/portal/protocol/push"
)

var (
	recvVal interface{}
	p0      = push.New(portal.Cfg{})
	p1      = pull.New(portal.Cfg{})
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

func benchPair(i int, b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		go p0.Send(n)
		recvVal = p1.Recv()
	}
}

func BenchmarkPair10(b *testing.B)         { benchPair(10, b) }
func BenchmarkPair100(b *testing.B)        { benchPair(100, b) }
func BenchmarkPair1000(b *testing.B)       { benchPair(1000, b) }
func BenchmarkPair10000(b *testing.B)      { benchPair(10000, b) }
func BenchmarkPair100000(b *testing.B)     { benchPair(100000, b) }
func BenchmarkPair1000000(b *testing.B)    { benchPair(1000000, b) }
func BenchmarkPair1000000000(b *testing.B) { benchPair(1000000000, b) }
