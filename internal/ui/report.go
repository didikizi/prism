package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/didikizi/prism/internal/gotest"
)

// Report renders the final screen: failure cards, a summary panel, and—when
// benchmarks ran—a benchmark section in the requested mode. width is the
// terminal column count; boxes are sized to fit within it.
func Report(run *gotest.Run, width int, benchMode BenchMode) string {
	w := clamp(width-2, 58, 98) // leave room for the 2-column border

	var b strings.Builder
	if fails := run.Failures(); len(fails) > 0 {
		fmt.Fprintf(&b, "\n%s\n\n", stFail.Bold(true).Render("  FAILURES"))
		for _, f := range fails {
			b.WriteString(failCard(f, w))
			b.WriteString("\n\n")
		}
	}
	b.WriteString(summary(run, w))
	b.WriteString("\n")

	if sec := benchSection(run.Benchmarks(), run.Env(), benchMode, width); sec != "" {
		b.WriteString("\n")
		b.WriteString(sec)
		b.WriteString("\n")
	}
	return b.String()
}

// packageLine is the permanent one-line result printed as a package finishes.
func packageLine(p *gotest.Package) string {
	name := stBold.Render(shorten(p.Name))
	dur := stTeal.Render(p.Elapsed.Round(time.Millisecond).String())

	switch p.Outcome {
	case gotest.Passed:
		note := stDim.Render(fmt.Sprintf("%d ok", p.Count(gotest.Passed)))
		return "  " + stPass.Render(glyphPass) + "  " + name + "  " + dur + "  " + note
	case gotest.Failed:
		nf := p.Count(gotest.Failed)
		note := stDim.Render("failed")
		if nf > 0 {
			note = stFail.Render(fmt.Sprintf("%d failed", nf))
		}
		return "  " + stFail.Render(glyphFail) + "  " + name + "  " + dur + "  " + note
	case gotest.Skipped:
		return "  " + stSkip.Render(glyphSkip) + "  " + stDim.Render(shorten(p.Name)) + "  " + stDim.Render("no tests")
	}
	return ""
}

// failCard renders one failure in a rounded box coloured by its kind.
func failCard(f gotest.Failure, w int) string {
	var border lipgloss.Color
	var tag string
	switch f.Kind {
	case gotest.Panic:
		border, tag = peach, badge("PANIC", peach)
	case gotest.Race:
		border, tag = yellow, badge("RACE", yellow)
	case gotest.Build:
		border, tag = red, badge("BUILD", red)
	default:
		border, tag = red, stFail.Bold(true).Render(glyphFail+" FAIL")
	}

	title := tag + "  " + stBold.Render(f.Test)
	sub := stDim.Render(f.Pkg)
	if loc := location(f.Output); loc != "" {
		sub += stDim.Render("  ·  ") + stTeal.Render(loc)
	}

	content := title + "\n" + sub
	if body := failBody(f, w-6); body != "" {
		content += "\n\n" + body
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Width(w).
		Render(content)
}

func failBody(f gotest.Failure, w int) string {
	switch f.Kind {
	case gotest.Panic:
		return panicBody(f.Output, w)
	case gotest.Race:
		return raceBody(f.Output, w)
	default:
		return assertionBody(f.Output, w)
	}
}

func assertionBody(out []string, w int) string {
	lines := stripNoise(out)
	for i := range lines {
		lines[i] = stripLocPrefix(lines[i])
	}
	lines = trimBlankEdges(dedent(lines))
	if len(lines) == 0 {
		return ""
	}
	rendered := make([]string, len(lines))
	for i, l := range clipLines(lines, w) {
		rendered[i] = stText.Render(l)
	}
	return strings.Join(rendered, "\n")
}

func panicBody(out []string, w int) string {
	msg, stack := splitPanic(out)
	stack = clipLines(truncate(stripNoise(stack), 12), w)

	var parts []string
	if msg != "" {
		parts = append(parts, stText.Render(clip(msg, w)))
	}
	if len(stack) > 0 {
		dim := make([]string, len(stack))
		for i, l := range stack {
			dim[i] = stFaint.Render(l)
		}
		parts = append(parts, strings.Join(dim, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

func raceBody(out []string, w int) string {
	lines := clipLines(truncate(stripNoise(out), 16), w)
	rendered := make([]string, len(lines))
	for i, l := range lines {
		if strings.Contains(l, "DATA RACE") {
			rendered[i] = lipgloss.NewStyle().Bold(true).Foreground(yellow).Render(l)
		} else {
			rendered[i] = stFaint.Render(l)
		}
	}
	return strings.Join(rendered, "\n")
}

// summary is the closing rounded panel: headline, stat row, slowest tests.
func summary(run *gotest.Run, w int) string {
	total := run.Pass + run.Fail + run.Skip
	elapsed := time.Since(run.Start).Round(time.Millisecond)

	border := green
	head := badge("PASS", green) + "  " + lipgloss.NewStyle().Bold(true).Foreground(green).Render("all tests passed")
	if run.Failed() {
		border = red
		head = badge("FAIL", red) + "  " + lipgloss.NewStyle().Bold(true).Foreground(red).Render(failLabel(run))
	}

	stat := stPass.Render(fmt.Sprintf("%s %d passed", glyphPass, run.Pass)) + "    " +
		failStat(run) + "    " +
		stSkip.Render(fmt.Sprintf("%s %d skipped", glyphSkip, run.Skip)) + "    " +
		stDim.Render("·") + "    " +
		stDim.Render(fmt.Sprintf("%s in %s", plural(total, "test"), elapsed))

	lines := []string{head, "", stat}
	if slow := run.Slowest(5); len(slow) > 0 {
		lines = append(lines, "", stDim.Render("slowest"))
		for i, t := range slow {
			lines = append(lines, slowRow(i+1, t))
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 3).
		Width(w).
		Render(strings.Join(lines, "\n"))
}

func failLabel(run *gotest.Run) string {
	if run.Fail > 0 {
		return fmt.Sprintf("%d failed", run.Fail)
	}
	return "build failed"
}

func failStat(run *gotest.Run) string {
	s := fmt.Sprintf("%s %d failed", glyphFail, run.Fail)
	if run.Fail > 0 {
		return stFail.Render(s)
	}
	return stDim.Render(s)
}

func slowRow(n int, t *gotest.Test) string {
	idx := stFaint.Render(fmt.Sprintf("%d", n))
	name := lipgloss.NewStyle().Width(34).Foreground(text).Render(clip(t.Name, 34))
	pkg := lipgloss.NewStyle().Width(14).Foreground(subtle).Render(clip(shorten(t.Pkg), 14))
	dur := stTeal.Render(t.Elapsed.Round(time.Millisecond).String())
	return "  " + idx + "  " + name + "  " + pkg + "  " + dur
}
