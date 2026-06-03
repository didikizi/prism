package gotest

import (
	"sort"
	"strings"
	"time"
)

// Outcome is the terminal state of a test or package.
type Outcome int

const (
	Running Outcome = iota
	Passed
	Failed
	Skipped
)

// FailKind classifies *why* a test or package failed, so the UI can render
// each kind distinctly without re-inspecting raw output.
type FailKind int

const (
	Assertion FailKind = iota // ordinary t.Error/t.Fatal
	Panic                     // runtime panic
	Race                      // -race data race
	Build                     // compile / build failure
)

// Test is one test (or subtest — names contain '/') within a package.
type Test struct {
	Name    string
	Pkg     string
	Outcome Outcome
	Elapsed time.Duration
	Output  []string
}

// Package collects the tests and package-level output of one import path.
type Package struct {
	Name    string
	Outcome Outcome
	Elapsed time.Duration
	Tests   []*Test
	Benches []*Benchmark // parsed `-bench` results, in first-seen order
	Output  []string     // package-level output (build errors, top-level panics)

	idx      map[string]*Test
	benchIdx map[string]*Benchmark
}

// Count returns how many of the package's tests ended with the given outcome.
func (p *Package) Count(o Outcome) int {
	n := 0
	for _, t := range p.Tests {
		if t.Outcome == o {
			n++
		}
	}
	return n
}

func (p *Package) test(name string) *Test {
	if t := p.idx[name]; t != nil {
		return t
	}
	t := &Test{Name: name, Pkg: p.Name}
	p.idx[name] = t
	p.Tests = append(p.Tests, t)
	return t
}

// addBench records a benchmark result, overwriting any earlier run of the same
// name (so -count=N keeps the last, most-settled measurement).
func (p *Package) addBench(b Benchmark) {
	b.Pkg = p.Name
	if existing := p.benchIdx[b.Name]; existing != nil {
		*existing = b
		return
	}
	nb := &b
	p.benchIdx[b.Name] = nb
	p.Benches = append(p.Benches, nb)
}

// Run is the accumulated state of an entire `go test` invocation.
type Run struct {
	Pass, Fail, Skip int
	Start            time.Time

	// Benchmark environment, parsed from the goos/goarch/cpu header lines that
	// `go test -bench` prints. Empty when no benchmarks ran.
	GOOS, GOARCH, CPU string

	pkgs   []*Package
	pkgIdx map[string]*Package

	// build-output lines accumulated by ImportPath, attached to a package when
	// its failure references them via FailedBuild.
	buildOut map[string][]string
}

// Env returns a one-line description of the benchmark hardware, or "" if none
// was reported (e.g. "linux/amd64 · Intel(R) Xeon(R) CPU E5-2697 v3 @ 2.60GHz").
func (r *Run) Env() string {
	plat := r.GOOS
	if r.GOARCH != "" {
		if plat != "" {
			plat += "/"
		}
		plat += r.GOARCH
	}
	switch {
	case plat != "" && r.CPU != "":
		return plat + " · " + r.CPU
	case plat != "":
		return plat
	default:
		return r.CPU
	}
}

// Failure is a flattened view of one thing that went wrong, ready to render.
type Failure struct {
	Test   string
	Pkg    string
	Kind   FailKind
	Output []string
}

// NewRun returns an empty Run with the clock started.
func NewRun() *Run {
	return &Run{Start: time.Now(), pkgIdx: map[string]*Package{}}
}

func (r *Run) pkg(name string) *Package {
	if p := r.pkgIdx[name]; p != nil {
		return p
	}
	p := &Package{Name: name, idx: map[string]*Test{}, benchIdx: map[string]*Benchmark{}}
	r.pkgIdx[name] = p
	r.pkgs = append(r.pkgs, p)
	return p
}

func seconds(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}

// failedBuildOutput resolves the compiler output for a failed build. It prefers
// the exact FailedBuild import path, then falls back to any captured build
// output whose import path starts with the package name (older toolchains that
// omit the FailedBuild field).
func (r *Run) failedBuildOutput(failedBuild, pkgName string) []string {
	if len(r.buildOut) == 0 {
		return nil
	}
	if failedBuild != "" {
		if out := r.buildOut[failedBuild]; len(out) > 0 {
			return out
		}
	}
	for path, out := range r.buildOut {
		if path == pkgName || strings.HasPrefix(path, pkgName+" ") {
			return out
		}
	}
	return nil
}

// Add applies one event. It returns the affected package and whether that
// package just reached a terminal state (so callers can print its result line).
func (r *Run) Add(ev Event) (pkg *Package, done bool) {
	// Build failures arrive before any package is created and are keyed by
	// ImportPath, not Package — handle them before touching the package map so
	// they don't spawn a phantom empty-named package.
	switch ev.Action {
	case "build-output":
		if r.buildOut == nil {
			r.buildOut = map[string][]string{}
		}
		r.buildOut[ev.ImportPath] = append(r.buildOut[ev.ImportPath], ev.Output)
		return nil, false
	case "build-fail":
		return nil, false
	}

	p := r.pkg(ev.Package)

	// Benchmarks emit run/output events but never pass/fail, so they are kept
	// out of the test model entirely; only their result line is captured.
	if ev.Test != "" && isBench(ev.Test) {
		if ev.Action == "output" {
			if b, ok := parseBench(ev.Output); ok {
				p.addBench(b)
			}
		}
		return p, false
	}

	switch ev.Action {
	case "run":
		if ev.Test != "" {
			p.test(ev.Test).Outcome = Running
		}

	case "output":
		switch {
		case ev.Test != "":
			t := p.test(ev.Test)
			t.Output = append(t.Output, ev.Output)
		case r.parseEnv(ev.Output):
			// consumed a goos/goarch/cpu header line
		default:
			if b, ok := parseBench(ev.Output); ok {
				p.addBench(b) // some toolchains attribute the result line to the package
			} else {
				p.Output = append(p.Output, ev.Output)
			}
		}

	case "pass":
		if ev.Test != "" {
			t := p.test(ev.Test)
			t.Outcome, t.Elapsed = Passed, seconds(ev.Elapsed)
			r.Pass++
		} else {
			p.Outcome, p.Elapsed = Passed, seconds(ev.Elapsed)
			done = true
		}

	case "fail":
		if ev.Test != "" {
			t := p.test(ev.Test)
			t.Outcome, t.Elapsed = Failed, seconds(ev.Elapsed)
			r.Fail++
		} else {
			p.Outcome, p.Elapsed = Failed, seconds(ev.Elapsed)
			if out := r.failedBuildOutput(ev.FailedBuild, p.Name); len(out) > 0 {
				p.Output = append(p.Output, out...)
			}
			done = true
		}

	case "skip":
		if ev.Test != "" {
			t := p.test(ev.Test)
			t.Outcome, t.Elapsed = Skipped, seconds(ev.Elapsed)
			r.Skip++
		} else {
			p.Outcome = Skipped
			done = true
		}
	}

	return p, done
}

// Packages returns the packages in the order they first appeared.
func (r *Run) Packages() []*Package { return r.pkgs }

// Benchmarks returns all parsed benchmark results in package/declaration order.
func (r *Run) Benchmarks() []*Benchmark {
	var out []*Benchmark
	for _, p := range r.pkgs {
		out = append(out, p.Benches...)
	}
	return out
}

// HasBenchmarks reports whether any benchmark results were seen.
func (r *Run) HasBenchmarks() bool {
	for _, p := range r.pkgs {
		if len(p.Benches) > 0 {
			return true
		}
	}
	return false
}

// Failed reports whether anything failed (a test, panic, race, or build).
func (r *Run) Failed() bool {
	if r.Fail > 0 {
		return true
	}
	for _, p := range r.pkgs {
		if p.Outcome == Failed {
			return true
		}
	}
	return false
}

// Failures returns every failure in package/declaration order. A package that
// failed with no individual failing test is reported once (panic/race/build).
func (r *Run) Failures() []Failure {
	var out []Failure
	for _, p := range r.pkgs {
		failedNames := map[string]bool{}
		for _, t := range p.Tests {
			if t.Outcome == Failed {
				failedNames[t.Name] = true
			}
		}
		hadTest := false
		for _, t := range p.Tests {
			if t.Outcome != Failed {
				continue
			}
			hadTest = true
			// Skip a parent test whose failure is just a roll-up of a failing
			// subtest — the leaf "Parent/Child" card carries the real detail.
			if hasFailingChild(t.Name, failedNames) {
				continue
			}
			out = append(out, Failure{
				Test:   t.Name,
				Pkg:    p.Name,
				Kind:   classify(t.Output),
				Output: t.Output,
			})
		}
		if p.Outcome == Failed && !hadTest {
			kind := classify(p.Output)
			if kind == Assertion {
				kind = Build // a package-level failure with no test is a build error
			}
			out = append(out, Failure{
				Test:   syntheticName(kind),
				Pkg:    p.Name,
				Kind:   kind,
				Output: p.Output,
			})
		}
	}
	return out
}

// Slowest returns up to n completed tests, slowest first.
func (r *Run) Slowest(n int) []*Test {
	var all []*Test
	for _, p := range r.pkgs {
		for _, t := range p.Tests {
			if t.Outcome == Passed || t.Outcome == Failed {
				all = append(all, t)
			}
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].Elapsed > all[j].Elapsed
	})
	if len(all) > n {
		all = all[:n]
	}
	return all
}

// classify inspects raw output to determine why something failed.
func classify(out []string) FailKind {
	for _, l := range out {
		if strings.Contains(l, "DATA RACE") {
			return Race
		}
	}
	for _, l := range out {
		if strings.HasPrefix(strings.TrimSpace(l), "panic:") {
			return Panic
		}
	}
	return Assertion
}

// hasFailingChild reports whether any failed test name is a subtest of name
// (i.e. starts with "name/").
func hasFailingChild(name string, failed map[string]bool) bool {
	prefix := name + "/"
	for n := range failed {
		if strings.HasPrefix(n, prefix) {
			return true
		}
	}
	return false
}

func syntheticName(k FailKind) string {
	switch k {
	case Panic:
		return "(panic)"
	case Race:
		return "(data race)"
	default:
		return "(build error)"
	}
}
