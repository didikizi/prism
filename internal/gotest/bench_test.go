package gotest

import "testing"

func TestParseBench(t *testing.T) {
	line := "BenchmarkAllocLarge-28    \t  195386650\t   11969 ns/op\t   32792 B/op\t       2 allocs/op\n"
	b, ok := parseBench(line)
	if !ok {
		t.Fatal("parseBench returned ok=false")
	}
	if b.Name != "BenchmarkAllocLarge" {
		t.Errorf("Name = %q, want BenchmarkAllocLarge (GOMAXPROCS suffix stripped)", b.Name)
	}
	if b.NsOp != 11969 {
		t.Errorf("NsOp = %v, want 11969", b.NsOp)
	}
	if b.Bytes != 32792 || b.Allocs != 2 || !b.HasMem {
		t.Errorf("mem = %d B / %d allocs / hasMem=%v", b.Bytes, b.Allocs, b.HasMem)
	}
}

func TestParseBenchNoMem(t *testing.T) {
	b, ok := parseBench("BenchmarkAdd-8 \t1000000000\t0.51 ns/op")
	if !ok {
		t.Fatal("ok=false")
	}
	if b.HasMem || b.Bytes != -1 || b.Allocs != -1 {
		t.Errorf("expected no mem stats, got %+v", b)
	}
}

func TestParseBenchRejectsNonBench(t *testing.T) {
	for _, line := range []string{
		"=== RUN   BenchmarkAdd",
		"BenchmarkAdd",
		"    calc_test.go:5: boom",
		"ok  \tpkg\t1.2s",
	} {
		if _, ok := parseBench(line); ok {
			t.Errorf("parseBench(%q) = ok, want rejected", line)
		}
	}
}

func TestBenchmarksExcludedFromTests(t *testing.T) {
	r := NewRun()
	feed(t, r,
		`{"Action":"run","Package":"x","Test":"BenchmarkAdd"}`,
		`{"Action":"output","Package":"x","Test":"BenchmarkAdd","Output":"BenchmarkAdd-8 \t1000000000\t0.51 ns/op\t0 B/op\t0 allocs/op\n"}`,
		`{"Action":"pass","Package":"x","Elapsed":1.0}`,
	)
	if got := len(r.Packages()[0].Tests); got != 0 {
		t.Errorf("benchmark leaked into Tests: %d entries", got)
	}
	bs := r.Benchmarks()
	if len(bs) != 1 || bs[0].Name != "BenchmarkAdd" {
		t.Fatalf("Benchmarks() = %+v, want one BenchmarkAdd", bs)
	}
	if !r.HasBenchmarks() {
		t.Error("HasBenchmarks() = false")
	}
}

func TestParseEnvHeader(t *testing.T) {
	r := NewRun()
	feed(t, r,
		`{"Action":"output","Package":"x","Output":"goos: linux\n"}`,
		`{"Action":"output","Package":"x","Output":"goarch: amd64\n"}`,
		`{"Action":"output","Package":"x","Output":"pkg: x\n"}`,
		`{"Action":"output","Package":"x","Output":"cpu: Intel(R) Xeon(R) CPU E5-2697 v3 @ 2.60GHz\n"}`,
	)
	if r.GOOS != "linux" || r.GOARCH != "amd64" {
		t.Errorf("platform = %s/%s, want linux/amd64", r.GOOS, r.GOARCH)
	}
	want := "linux/amd64 · Intel(R) Xeon(R) CPU E5-2697 v3 @ 2.60GHz"
	if got := r.Env(); got != want {
		t.Errorf("Env() = %q, want %q", got, want)
	}
	// header lines must not leak into the package output (build-error) bucket
	if got := len(r.Packages()[0].Output); got != 0 {
		t.Errorf("env headers leaked into package output: %d lines", got)
	}
}

func TestAddBenchDedupesByName(t *testing.T) {
	r := NewRun()
	feed(t, r,
		`{"Action":"output","Package":"x","Test":"BenchmarkAdd","Output":"BenchmarkAdd-8 \t100\t1.0 ns/op\n"}`,
		`{"Action":"output","Package":"x","Test":"BenchmarkAdd","Output":"BenchmarkAdd-8 \t200\t2.0 ns/op\n"}`,
	)
	bs := r.Benchmarks()
	if len(bs) != 1 {
		t.Fatalf("len = %d, want 1 (deduped)", len(bs))
	}
	if bs[0].NsOp != 2.0 {
		t.Errorf("NsOp = %v, want 2.0 (last wins)", bs[0].NsOp)
	}
}
