package model

import (
	"sort"
	"strings"
	"time"

	"github.com/didikizi/prism/internal/parser"
)

type Status int

const (
	StatusRunning Status = iota
	StatusPassed
	StatusFailed
	StatusSkipped
)

type Test struct {
	Name    string
	Package string
	Status  Status
	Elapsed float64
	Output  []string
	IsPanic bool
}

type Package struct {
	Name    string
	Tests   map[string]*Test
	Output  []string // package-level output (build errors, panics)
	Passed  int
	Failed  int
	Skipped int
	Elapsed float64
	Result  string // "pass" | "fail" | "skip"
	IsPanic bool
}

type State struct {
	Packages  map[string]*Package
	pkgOrder  []string
	Passed    int
	Failed    int
	Skipped   int
	StartTime time.Time
}

func New() *State {
	return &State{
		Packages:  make(map[string]*Package),
		StartTime: time.Now(),
	}
}

func (s *State) pkg(name string) *Package {
	if _, ok := s.Packages[name]; !ok {
		s.Packages[name] = &Package{
			Name:  name,
			Tests: make(map[string]*Test),
		}
		s.pkgOrder = append(s.pkgOrder, name)
	}
	return s.Packages[name]
}

func (s *State) Apply(ev *parser.TestEvent) {
	pkg := s.pkg(ev.Package)

	switch ev.Action {
	case "run":
		if ev.Test != "" {
			pkg.Tests[ev.Test] = &Test{
				Name:    ev.Test,
				Package: ev.Package,
				Status:  StatusRunning,
			}
		}

	case "output":
		if ev.Test != "" {
			if t, ok := pkg.Tests[ev.Test]; ok {
				t.Output = append(t.Output, ev.Output)
				if strings.Contains(ev.Output, "panic:") {
					t.IsPanic = true
				}
			}
		} else {
			pkg.Output = append(pkg.Output, ev.Output)
			if strings.Contains(ev.Output, "panic:") {
				pkg.IsPanic = true
			}
		}

	case "pass":
		if ev.Test != "" {
			if t, ok := pkg.Tests[ev.Test]; ok {
				t.Status = StatusPassed
				t.Elapsed = ev.Elapsed
				pkg.Passed++
				s.Passed++
			}
		} else {
			pkg.Result = "pass"
			pkg.Elapsed = ev.Elapsed
		}

	case "fail":
		if ev.Test != "" {
			if t, ok := pkg.Tests[ev.Test]; ok {
				t.Status = StatusFailed
				t.Elapsed = ev.Elapsed
				pkg.Failed++
				s.Failed++
			}
		} else {
			pkg.Result = "fail"
			pkg.Elapsed = ev.Elapsed
			// A package that failed with no test failures is a panic or build error.
			if pkg.Failed == 0 && pkg.IsPanic {
				s.Failed++ // count it so exit code is non-zero
			}
		}

	case "skip":
		if ev.Test != "" {
			if t, ok := pkg.Tests[ev.Test]; ok {
				t.Status = StatusSkipped
				t.Elapsed = ev.Elapsed
				pkg.Skipped++
				s.Skipped++
			}
		} else {
			pkg.Result = "skip"
		}
	}
}

func (s *State) HasFailures() bool {
	return s.Failed > 0
}

// FailedTests returns all failed tests plus a synthetic entry for panicked packages.
func (s *State) FailedTests() []*Test {
	var out []*Test
	for _, name := range s.pkgOrder {
		pkg := s.Packages[name]
		if pkg.IsPanic && pkg.Failed == 0 {
			out = append(out, &Test{
				Name:    "(package panic)",
				Package: pkg.Name,
				Status:  StatusFailed,
				Output:  pkg.Output,
				IsPanic: true,
			})
		}
		for _, t := range pkg.Tests {
			if t.Status == StatusFailed {
				out = append(out, t)
			}
		}
	}
	return out
}

// SlowestTests returns up to n tests sorted by elapsed time descending.
func (s *State) SlowestTests(n int) []*Test {
	var all []*Test
	for _, pkg := range s.Packages {
		for _, t := range pkg.Tests {
			if t.Status == StatusPassed || t.Status == StatusFailed {
				all = append(all, t)
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Elapsed > all[j].Elapsed
	})
	if len(all) > n {
		return all[:n]
	}
	return all
}
