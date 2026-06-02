package ui

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// matches a test-file reference like "calc_test.go:25" or "/abs/x_test.go:9"
	locRe = regexp.MustCompile(`[\w.\-/]+_test\.go:\d+`)
	// matches a leading "    file_test.go:25: " location prefix on a line
	locPrefixRe = regexp.MustCompile(`^\s*[\w.\-/]+_test\.go:\d+:\s?`)
)

// location returns the first "file:line" found in output, as a bare basename.
func location(out []string) string {
	for _, l := range out {
		if m := locRe.FindString(l); m != "" {
			return filepath.Base(m)
		}
	}
	return ""
}

func stripLocPrefix(line string) string {
	return locPrefixRe.ReplaceAllString(line, "")
}

// stripNoise drops go test framework chatter, keeping only meaningful lines.
func stripNoise(out []string) []string {
	var keep []string
	for _, raw := range out {
		l := strings.TrimRight(raw, "\r\n")
		t := strings.TrimSpace(l)
		switch {
		case t == "", t == "PASS", t == "FAIL":
			continue
		case strings.Trim(t, "=") == "": // "======" separators
			continue
		case strings.HasPrefix(t, "=== "),
			strings.HasPrefix(t, "--- FAIL"),
			strings.HasPrefix(t, "--- PASS"),
			strings.HasPrefix(t, "--- SKIP"),
			strings.HasPrefix(t, "ok  "),
			strings.HasPrefix(t, "FAIL\t"),
			strings.HasPrefix(t, "exit status"):
			continue
		}
		keep = append(keep, l)
	}
	return keep
}

// dedent removes the common leading indentation shared by all non-blank lines.
func dedent(lines []string) []string {
	min := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " \t"))
		if min < 0 || n < min {
			min = n
		}
	}
	if min <= 0 {
		return lines
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		if len(l) >= min {
			out[i] = l[min:]
		} else {
			out[i] = strings.TrimLeft(l, " \t")
		}
	}
	return out
}

func trimBlankEdges(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// truncate caps lines to max, appending a "… N more" marker when it cuts.
func truncate(lines []string, max int) []string {
	if len(lines) <= max {
		return lines
	}
	out := append([]string(nil), lines[:max]...)
	return append(out, "… "+strconv.Itoa(len(lines)-max)+" more")
}

func clipLines(lines []string, w int) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = clip(l, w)
	}
	return out
}

// splitPanic separates the "panic: ..." headline from the stack trace below it.
func splitPanic(out []string) (msg string, stack []string) {
	for i, raw := range out {
		t := strings.TrimSpace(raw)
		if strings.HasPrefix(t, "panic:") {
			return t, out[i+1:]
		}
	}
	return "", out
}
