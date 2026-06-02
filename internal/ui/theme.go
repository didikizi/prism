package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha — jewel tones that photograph well on a dark terminal.
const (
	green  lipgloss.Color = "#a6e3a1"
	red    lipgloss.Color = "#f38ba8"
	yellow lipgloss.Color = "#f9e2af"
	peach  lipgloss.Color = "#fab387"
	teal   lipgloss.Color = "#94e2d5"
	mauve  lipgloss.Color = "#cba6f7"
	text   lipgloss.Color = "#cdd6f4"
	subtle lipgloss.Color = "#6c7086"
	faint  lipgloss.Color = "#45475a"
	base   lipgloss.Color = "#11111b"
)

const (
	glyphPass = "✓"
	glyphFail = "✗"
	glyphSkip = "⊘"
)

var (
	stPass  = lipgloss.NewStyle().Foreground(green)
	stFail  = lipgloss.NewStyle().Foreground(red)
	stSkip  = lipgloss.NewStyle().Foreground(subtle)
	stDim   = lipgloss.NewStyle().Foreground(subtle)
	stFaint = lipgloss.NewStyle().Foreground(faint)
	stText  = lipgloss.NewStyle().Foreground(text)
	stBold  = lipgloss.NewStyle().Bold(true).Foreground(text)
	stTeal  = lipgloss.NewStyle().Foreground(teal)
	stSpin  = lipgloss.NewStyle().Foreground(mauve)
)

// badge renders a solid pill, e.g. "PANIC" on a peach background.
func badge(label string, bg lipgloss.Color) string {
	return lipgloss.NewStyle().Bold(true).Background(bg).Foreground(base).Padding(0, 1).Render(label)
}

// shorten reduces an import path to its final segment for compact display.
func shorten(pkg string) string {
	if i := strings.LastIndexByte(pkg, '/'); i >= 0 {
		return pkg[i+1:]
	}
	return pkg
}

// plural formats a count with its noun, adding "s" unless the count is 1.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clip truncates s to a display width of w, appending an ellipsis if cut.
func clip(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		return "…"
	}
	r := []rune(s)
	if len(r) > w-1 {
		r = r[:w-1]
	}
	return string(r) + "…"
}
