package benchchan

import (
	"testing"
)

var (
	recvVal int
	ch      = make(chan int)
)

func bench(i int, b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		go func() { ch <- n }()
		recvVal = <-ch
	}
}

func Benchmark10(b *testing.B)         { bench(10, b) }
func Benchmark100(b *testing.B)        { bench(100, b) }
func Benchmark1000(b *testing.B)       { bench(1000, b) }
func Benchmark10000(b *testing.B)      { bench(10000, b) }
func Benchmark100000(b *testing.B)     { bench(100000, b) }
func Benchmark1000000(b *testing.B)    { bench(1000000, b) }
func Benchmark1000000000(b *testing.B) { bench(1000000000, b) }
