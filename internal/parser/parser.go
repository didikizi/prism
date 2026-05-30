package parser

import (
	"encoding/json"
	"time"
)

// TestEvent mirrors the fields emitted by cmd/test2json.
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

// ParseLine decodes a single line of go test -json output.
// Returns an error for any non-JSON line (e.g. build output).
func ParseLine(line string) (*TestEvent, error) {
	var ev TestEvent
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
