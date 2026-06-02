package calc

import (
	"sort"
	"strconv"
	"testing"
)

// Benchmarks of varying cost and allocation so the relative-speed bars and the
// B/op column have a nice spread. Run them through prism with: make demo-bench
//
// Results are stored in sink to keep the compiler from optimising the work away.
var sink any

func BenchmarkAdd(b *testing.B) {
	x := 0
	for i := 0; i < b.N; i++ {
		x = add(x, i)
	}
	sink = x
}

func BenchmarkItoa(b *testing.B) {
	var s string
	for i := 0; i < b.N; i++ {
		s = strconv.Itoa(i)
	}
	sink = s
}

func BenchmarkSortSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xs := []int{5, 3, 8, 1, 9, 2, 7, 4, 6, 0}
		sort.Ints(xs)
		sink = xs
	}
}

func BenchmarkAllocLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xs := make([]int, 4096)
		xs[0] = i
		sink = xs
	}
}
