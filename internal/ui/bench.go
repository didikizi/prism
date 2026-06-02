package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/didikizi/prism/internal/gotest"
)

// BenchMode selects how benchmark results are rendered.
type BenchMode int

const (
	BenchBoth     BenchMode = iota // styled table + copyable markdown (default)
	BenchStyled                    // styled table only — screenshot-first
	BenchMarkdown                  // markdown table only — paste into a project
)

// ParseBenchMode maps a flag value to a BenchMode.
func ParseBenchMode(s string) (BenchMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "both", "":
		return BenchBoth, nil
	case "styled", "screenshot", "term":
		return BenchStyled, nil
	case "md", "markdown", "copy":
		return BenchMarkdown, nil
	default:
		return BenchBoth, fmt.Errorf("invalid --bench value %q (want both|styled|md)", s)
	}
}

// benchSection renders the benchmark block according to mode, or "" if there
// are no benchmarks. env is a one-line hardware description (may be empty).
func benchSection(benches []*gotest.Benchmark, env string, mode BenchMode, w int) string {
	if len(benches) == 0 {
		return ""
	}
	switch mode {
	case BenchStyled:
		return benchStyled(benches, env, w)
	case BenchMarkdown:
		return benchMarkdown(benches, env)
	default: // BenchBoth
		return benchStyled(benches, env, w) + "\n\n" +
			stDim.Render("copy ↓") + "\n" +
			benchMarkdown(benches, env)
	}
}

// benchStyled renders a rounded panel: name, humanized ns/op, a relative-speed
// bar (coloured fastest→slowest), and memory stats when present.
func benchStyled(benches []*gotest.Benchmark, env string, w int) string {
	inner := clamp(w-2, 58, 98)
	var maxNs float64
	hasMem := false
	for _, b := range benches {
		if b.NsOp > maxNs {
			maxNs = b.NsOp
		}
		hasMem = hasMem || b.HasMem
	}

	const barCells = 12
	rows := make([]string, 0, len(benches))
	for _, b := range benches {
		name := lipgloss.NewStyle().Width(20).Foreground(text).Render(clip(benchShort(b.Name), 20))
		ns := lipgloss.NewStyle().Width(11).Align(lipgloss.Right).Foreground(teal).Render(humanNs(b.NsOp))

		frac := 0.0
		if maxNs > 0 {
			frac = b.NsOp / maxNs
		}
		bar := lipgloss.NewStyle().Width(barCells).Foreground(barColor(frac)).Render(barGlyphs(frac, barCells))

		row := "  " + name + "  " + ns + "   " + bar
		if hasMem {
			mem := stDim.Render(fmt.Sprintf("%8s   %s", bytesHuman(b.Bytes), allocsHuman(b.Allocs)))
			row += "  " + mem
		}
		rows = append(rows, row)
	}

	header := stBold.Render("benchmarks")
	if env != "" {
		header += "\n" + stDim.Render(clip(env, inner-4))
	}
	content := header + "\n\n" + strings.Join(rows, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mauve).
		Padding(1, 2).
		Width(inner).
		Render(content)
}

// benchMarkdown renders a GitHub-flavoured table, column-padded so the raw text
// is tidy too. No ANSI — safe to copy straight into a README or PR.
func benchMarkdown(benches []*gotest.Benchmark, env string) string {
	hasMem := false
	for _, b := range benches {
		hasMem = hasMem || b.HasMem
	}

	headers := []string{"Benchmark", "ns/op"}
	aligns := []bool{false, true} // right-aligned?
	if hasMem {
		headers = append(headers, "B/op", "allocs/op")
		aligns = append(aligns, true, true)
	}

	rows := [][]string{}
	for _, b := range benches {
		cells := []string{b.Name, nsExact(b.NsOp)}
		if hasMem {
			cells = append(cells, int64Cell(b.Bytes), int64Cell(b.Allocs))
		}
		rows = append(rows, cells)
	}

	// column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}

	var b strings.Builder
	if env != "" {
		// italic caption above the table — renders as muted text on GitHub
		fmt.Fprintf(&b, "_%s_\n\n", env)
	}
	writeRow := func(cells []string) {
		b.WriteString("|")
		for i, c := range cells {
			if aligns[i] {
				fmt.Fprintf(&b, " %*s |", widths[i], c)
			} else {
				fmt.Fprintf(&b, " %-*s |", widths[i], c)
			}
		}
		b.WriteString("\n")
	}

	writeRow(headers)
	b.WriteString("|")
	for i := range headers {
		if aligns[i] {
			b.WriteString(" " + strings.Repeat("-", widths[i]-1) + ": |")
		} else {
			b.WriteString(" " + strings.Repeat("-", widths[i]) + " |")
		}
	}
	b.WriteString("\n")
	for _, r := range rows {
		writeRow(r)
	}
	return strings.TrimRight(b.String(), "\n")
}

func benchShort(name string) string {
	if s := strings.TrimPrefix(name, "Benchmark"); s != "" {
		return s
	}
	return name
}

// humanNs formats nanoseconds with an adaptive unit and two decimals.
func humanNs(ns float64) string {
	switch {
	case ns < 1e3:
		return fmt.Sprintf("%.2f ns", ns)
	case ns < 1e6:
		return fmt.Sprintf("%.2f µs", ns/1e3)
	case ns < 1e9:
		return fmt.Sprintf("%.2f ms", ns/1e6)
	default:
		return fmt.Sprintf("%.2f s", ns/1e9)
	}
}

// nsExact formats ns/op for markdown: full integers, decimals only when small.
func nsExact(ns float64) string {
	switch {
	case ns >= 100:
		return strconv.FormatFloat(ns, 'f', 0, 64)
	case ns >= 1:
		return strconv.FormatFloat(ns, 'f', 2, 64)
	default:
		return strconv.FormatFloat(ns, 'f', 4, 64)
	}
}

func int64Cell(v int64) string {
	if v < 0 {
		return "—"
	}
	return strconv.FormatInt(v, 10)
}

func bytesHuman(v int64) string {
	if v < 0 {
		return ""
	}
	if v < 1024 {
		return fmt.Sprintf("%d B", v)
	}
	return fmt.Sprintf("%.1f kB", float64(v)/1024)
}

func allocsHuman(v int64) string {
	if v < 0 {
		return ""
	}
	if v == 1 {
		return "1 alloc"
	}
	return fmt.Sprintf("%d allocs", v)
}

// barColor shades the bar fastest (green) → mid (yellow) → slowest (peach).
func barColor(frac float64) lipgloss.Color {
	switch {
	case frac < 0.34:
		return green
	case frac < 0.67:
		return yellow
	default:
		return peach
	}
}

// barGlyphs draws a proportional bar using eighth-block characters.
func barGlyphs(frac float64, cells int) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	units := frac * float64(cells)
	full := int(units)
	s := strings.Repeat("█", full)
	if rem := units - float64(full); full < cells && rem > 0 {
		eighths := []rune("▏▎▍▌▋▊▉█")
		idx := int(rem * 8)
		if idx > 7 {
			idx = 7
		}
		s += string(eighths[idx])
	}
	if s == "" {
		s = "▏" // always show a sliver so the fastest row isn't blank
	}
	return s
}
