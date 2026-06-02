// Package gotest decodes a `go test -json` event stream and aggregates it
// into a model of the run: packages, tests, outcomes, and classified failures.
package gotest

import "encoding/json"

// Event is a single cmd/test2json record emitted by `go test -json`.
type Event struct {
	Action  string  `json:"Action"`  // run, output, pass, fail, skip, ...
	Package string  `json:"Package"`
	Test    string  `json:"Test"`    // empty for package-level events
	Elapsed float64 `json:"Elapsed"` // seconds
	Output  string  `json:"Output"`
}

// Decode parses one line of the stream. ok is false for non-JSON lines
// (build chatter, blank lines), which callers should skip.
func Decode(line string) (ev Event, ok bool) {
	if line == "" || line[0] != '{' {
		return Event{}, false
	}
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return Event{}, false
	}
	return ev, true
}
