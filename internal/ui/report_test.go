package ui

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/didikizi/prism/internal/gotest"
	"github.com/muesli/termenv"
)

var update = flag.Bool("update", false, "update .golden files")

// TestMain forces a colour-free profile so golden output is deterministic and
// independent of whether tests run under a TTY.
func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii)
	os.Exit(m.Run())
}

// the summary prints wall-clock elapsed ("N tests in 1ms"); blank it out.
var elapsedRe = regexp.MustCompile(`(tests? in )[0-9.]+(?:ns|µs|ms|s)`)

func normalize(s string) string {
	return elapsedRe.ReplaceAllString(s, `${1}<elapsed>`)
}

func runFromJSONL(jsonl string) *gotest.Run {
	r := gotest.NewRun()
	for _, line := range strings.Split(jsonl, "\n") {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		if ev, ok := gotest.Decode(line); ok {
			r.Add(ev)
		}
	}
	return r
}

// TestReportGolden renders each testdata/*.jsonl fixture at a fixed width and
// compares it to the matching .golden file. Regenerate with:
//
//	go test ./internal/ui -run TestReportGolden -update
func TestReportGolden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no testdata/*.jsonl fixtures found")
	}

	for _, in := range inputs {
		name := strings.TrimSuffix(filepath.Base(in), ".jsonl")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(in)
			if err != nil {
				t.Fatal(err)
			}
			run := runFromJSONL(string(data))
			got := normalize(Report(run, 80, BenchBoth))

			golden := strings.TrimSuffix(in, ".jsonl") + ".golden"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run with -update to create): %v", err)
			}
			if got != string(want) {
				t.Errorf("output mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
			}
		})
	}
}
