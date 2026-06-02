package gotest

import (
	"regexp"
	"strconv"
	"strings"
)

// Benchmark is one parsed `go test -bench` result line.
//
// Benchmark numbers are not part of the structured JSON — `go test -json` only
// emits them as text inside output events — so this is the one place prism reads
// human-formatted output, via a single bounded regex.
type Benchmark struct {
	Name   string // full name, e.g. "BenchmarkRender"
	Pkg    string
	NsOp   float64
	Bytes  int64 // B/op
	Allocs int64 // allocs/op
	HasMem bool  // whether -benchmem stats are present
}

var (
	// e.g. "BenchmarkRender-8 \t195386650\t  6.163 ns/op"
	benchRe  = regexp.MustCompile(`^(Benchmark[^\s]*?)(?:-\d+)?\s+\d+\s+([0-9.]+)\s+ns/op`)
	bytesRe  = regexp.MustCompile(`([0-9]+)\s+B/op`)
	allocsRe = regexp.MustCompile(`([0-9]+)\s+allocs/op`)
)

// isBench reports whether a test name denotes a benchmark.
func isBench(name string) bool { return strings.HasPrefix(name, "Benchmark") }

// parseEnv consumes a `go test -bench` header line (goos/goarch/cpu), recording
// it on the run. It returns true when the line was such a header.
func (r *Run) parseEnv(line string) bool {
	t := strings.TrimRight(line, "\r\n")
	switch {
	case strings.HasPrefix(t, "goos: "):
		r.GOOS = strings.TrimSpace(t[len("goos: "):])
	case strings.HasPrefix(t, "goarch: "):
		r.GOARCH = strings.TrimSpace(t[len("goarch: "):])
	case strings.HasPrefix(t, "cpu: "):
		r.CPU = strings.TrimSpace(t[len("cpu: "):])
	case strings.HasPrefix(t, "pkg: "):
		// package path — already tracked per-Package; consume it silently
	default:
		return false
	}
	return true
}

// parseBench extracts a Benchmark from one output line, if it is a result line.
func parseBench(line string) (Benchmark, bool) {
	m := benchRe.FindStringSubmatch(line)
	if m == nil {
		return Benchmark{}, false
	}
	ns, err := strconv.ParseFloat(m[2], 64)
	if err != nil {
		return Benchmark{}, false
	}
	b := Benchmark{Name: m[1], NsOp: ns, Bytes: -1, Allocs: -1}
	if bm := bytesRe.FindStringSubmatch(line); bm != nil {
		if v, err := strconv.ParseInt(bm[1], 10, 64); err == nil {
			b.Bytes, b.HasMem = v, true
		}
	}
	if am := allocsRe.FindStringSubmatch(line); am != nil {
		if v, err := strconv.ParseInt(am[1], 10, 64); err == nil {
			b.Allocs, b.HasMem = v, true
		}
	}
	return b, true
}
