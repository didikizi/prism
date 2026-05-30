package render

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha palette — jewel tones that photograph well.
var (
	colorPass    = lipgloss.Color("#a6e3a1") // green
	colorFail    = lipgloss.Color("#f38ba8") // red
	colorSkip    = lipgloss.Color("#6c7086") // overlay0
	colorPkg     = lipgloss.Color("#89b4fa") // blue
	colorTime    = lipgloss.Color("#94e2d5") // teal
	colorMuted   = lipgloss.Color("#585b70") // surface2
	colorSpinner = lipgloss.Color("#cba6f7") // mauve
	colorPanic   = lipgloss.Color("#fab387") // peach
	colorText    = lipgloss.Color("#cdd6f4") // text
	colorDark    = lipgloss.Color("#1e1e2e") // base
)

type styles struct {
	passIcon     lipgloss.Style
	failIcon     lipgloss.Style
	skipIcon     lipgloss.Style
	pass         lipgloss.Style
	fail         lipgloss.Style
	skip         lipgloss.Style
	pkgName      lipgloss.Style
	elapsed      lipgloss.Style
	muted        lipgloss.Style
	spinner      lipgloss.Style
	outputText   lipgloss.Style
	failTitle    lipgloss.Style
	panicTitle   lipgloss.Style
	panicBadge   lipgloss.Style
	sectionTitle lipgloss.Style
	passColor    lipgloss.Color
	failColor    lipgloss.Color
	panicColor   lipgloss.Color
}

func newStyles() *styles {
	return &styles{
		passIcon:     lipgloss.NewStyle().Foreground(colorPass).Bold(true),
		failIcon:     lipgloss.NewStyle().Foreground(colorFail).Bold(true),
		skipIcon:     lipgloss.NewStyle().Foreground(colorSkip),
		pass:         lipgloss.NewStyle().Foreground(colorPass),
		fail:         lipgloss.NewStyle().Foreground(colorFail),
		skip:         lipgloss.NewStyle().Foreground(colorSkip),
		pkgName:      lipgloss.NewStyle().Foreground(colorPkg),
		elapsed:      lipgloss.NewStyle().Foreground(colorTime),
		muted:        lipgloss.NewStyle().Foreground(colorMuted),
		spinner:      lipgloss.NewStyle().Foreground(colorSpinner),
		outputText:   lipgloss.NewStyle().Foreground(colorText),
		failTitle:    lipgloss.NewStyle().Bold(true).Foreground(colorFail),
		panicTitle:   lipgloss.NewStyle().Bold(true).Foreground(colorPanic),
		panicBadge:   lipgloss.NewStyle().Bold(true).Background(colorPanic).Foreground(colorDark).Padding(0, 1),
		sectionTitle: lipgloss.NewStyle().Bold(true).Foreground(colorFail).Padding(0, 1),
		passColor:    colorPass,
		failColor:    colorFail,
		panicColor:   colorPanic,
	}
}
