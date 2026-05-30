package render

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
	"github.com/didikizi/prism/internal/model"
	"github.com/didikizi/prism/internal/parser"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Renderer struct {
	state *model.State
	st    *styles
	width int
	out   *os.File

	mu         sync.Mutex
	wg         sync.WaitGroup
	spinnerIdx int
	stopCh     chan struct{}
	atLine     bool // true when a non-newline-terminated status line is on screen
	lineLen    int  // visible width of the current status line
}

func New(state *model.State) *Renderer {
	return &Renderer{
		state:  state,
		st:     newStyles(),
		width:  termWidth(),
		out:    os.Stdout,
		stopCh: make(chan struct{}),
	}
}

// Start launches the background spinner goroutine.
func (r *Renderer) Start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				r.mu.Lock()
				r.spinnerIdx = (r.spinnerIdx + 1) % len(spinnerFrames)
				r.redrawStatus()
				r.mu.Unlock()
			case <-r.stopCh:
				return
			}
		}
	}()
}

// HandleEvent processes one test2json event and updates output as needed.
func (r *Renderer) HandleEvent(ev *parser.TestEvent) {
	r.state.Apply(ev)
	// Print a permanent line when a package finishes.
	if ev.Test == "" && (ev.Action == "pass" || ev.Action == "fail" || ev.Action == "skip") {
		r.mu.Lock()
		r.printPkgLine(ev.Package)
		r.mu.Unlock()
	}
}

// Finish stops the spinner and renders the final summary.
func (r *Renderer) Finish() {
	close(r.stopCh)
	r.wg.Wait()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.clearLine()
	fmt.Fprintln(r.out)

	failed := r.state.FailedTests()
	if len(failed) > 0 {
		fmt.Fprintln(r.out, r.st.sectionTitle.Render(" FAILURES "))
		fmt.Fprintln(r.out)
		for _, t := range failed {
			fmt.Fprintln(r.out, r.renderFailCard(t))
		}
	}

	fmt.Fprintln(r.out, r.renderSummary())
}

// --- internal helpers ---

// clearLine erases the current status line (no-op if not at a status line).
func (r *Renderer) clearLine() {
	if r.atLine {
		fmt.Fprintf(r.out, "\r%s\r", strings.Repeat(" ", r.lineLen+2))
		r.atLine = false
		r.lineLen = 0
	}
}

// redrawStatus clears and re-renders the spinner/counter line.
func (r *Renderer) redrawStatus() {
	r.clearLine()
	frame := spinnerFrames[r.spinnerIdx]
	s := r.st.spinner.Render(frame) + "  " +
		r.st.muted.Render("testing") + "   " +
		r.st.passIcon.Render("✓") + " " + r.st.pass.Render(fmt.Sprintf("%d", r.state.Passed)) + "   " +
		r.st.failIcon.Render("✗") + " " + r.st.fail.Render(fmt.Sprintf("%d", r.state.Failed)) + "   " +
		r.st.skipIcon.Render("⊘") + " " + r.st.skip.Render(fmt.Sprintf("%d", r.state.Skipped))
	fmt.Fprint(r.out, s)
	r.atLine = true
	r.lineLen = lipgloss.Width(s)
}

// writeln clears the status line, writes s + newline, then leaves the line clean.
func (r *Renderer) writeln(s string) {
	r.clearLine()
	fmt.Fprintln(r.out, s)
}

func (r *Renderer) printPkgLine(pkgName string) {
	pkg, ok := r.state.Packages[pkgName]
	if !ok || pkg.Result == "" {
		return
	}
	short := shorten(pkgName)
	el := r.st.elapsed.Render(fmt.Sprintf("%.3fs", pkg.Elapsed))

	var line string
	switch pkg.Result {
	case "pass":
		line = r.st.passIcon.Render("✓") + "  " + r.st.pkgName.Render(short) + "  " + el
	case "fail":
		line = r.st.failIcon.Render("✗") + "  " + r.st.pkgName.Render(short) + "  " + el
	case "skip":
		line = r.st.skipIcon.Render("⊘") + "  " + r.st.muted.Render(short) + "  " + r.st.muted.Render("(no test files)")
	}
	r.writeln(line)
}

func (r *Renderer) renderFailCard(t *model.Test) string {
	var titleLine string
	if t.IsPanic {
		titleLine = r.st.panicBadge.Render(" PANIC ") + "  " + r.st.panicTitle.Render(t.Name)
	} else {
		titleLine = r.st.failIcon.Render("✗") + "  " + r.st.failTitle.Render(t.Name)
	}
	pkgLine := r.st.muted.Render(t.Package)

	filtered := filterOutput(t.Output)
	var body string
	if len(filtered) > 0 {
		joined := strings.TrimRight(strings.Join(filtered, ""), "\n")
		body = r.st.outputText.Render(joined)
	}

	var content string
	if body != "" {
		content = titleLine + "\n" + pkgLine + "\n\n" + body
	} else {
		content = titleLine + "\n" + pkgLine
	}

	borderColor := r.st.failColor
	if t.IsPanic {
		borderColor = r.st.panicColor
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(r.width - 4).
		Render(content)
}

func (r *Renderer) renderSummary() string {
	elapsed := time.Since(r.state.StartTime)
	total := r.state.Passed + r.state.Failed + r.state.Skipped

	var failStr string
	if r.state.Failed > 0 {
		failStr = r.st.fail.Render(fmt.Sprintf("%d failed", r.state.Failed))
	} else {
		failStr = r.st.muted.Render(fmt.Sprintf("%d failed", r.state.Failed))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		r.st.passIcon.Render("✓")+"  "+r.st.pass.Render(fmt.Sprintf("%d passed", r.state.Passed))+"     ",
		r.st.failIcon.Render("✗")+"  "+failStr+"     ",
		r.st.skipIcon.Render("⊘")+"  "+r.st.skip.Render(fmt.Sprintf("%d skipped", r.state.Skipped))+"     ",
		r.st.muted.Render(fmt.Sprintf("%d total  %.2fs", total, elapsed.Seconds())),
	)

	var sections []string
	sections = append(sections, row)

	slowest := r.state.SlowestTests(5)
	if len(slowest) > 0 {
		sections = append(sections, "")
		sections = append(sections, r.st.muted.Render("slowest"))
		for i, t := range slowest {
			num := r.st.muted.Render(fmt.Sprintf("%d.", i+1))
			dur := r.st.elapsed.Render(fmt.Sprintf("%.3fs", t.Elapsed))
			pkg := r.st.muted.Render(shorten(t.Package))
			name := t.Name
			sections = append(sections, fmt.Sprintf("  %s  %-42s  %s  %s", num, name, pkg, dur))
		}
	}

	borderColor := r.st.passColor
	if r.state.Failed > 0 {
		borderColor = r.st.failColor
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 3).
		Width(r.width - 2).
		Render(strings.Join(sections, "\n"))
}

// filterOutput strips test framework noise, keeping only meaningful lines.
func filterOutput(lines []string) []string {
	var out []string
	for _, l := range lines {
		s := strings.TrimSpace(l)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "=== RUN") ||
			strings.HasPrefix(s, "=== PAUSE") ||
			strings.HasPrefix(s, "=== CONT") ||
			strings.HasPrefix(s, "--- PASS") ||
			strings.HasPrefix(s, "--- FAIL") ||
			strings.HasPrefix(s, "--- SKIP") ||
			s == "PASS" || s == "FAIL" ||
			strings.HasPrefix(s, "ok  \t") ||
			strings.HasPrefix(s, "FAIL\t") {
			continue
		}
		out = append(out, l)
	}
	return out
}

// shorten trims a long package path to its last two components.
func shorten(pkg string) string {
	parts := strings.Split(pkg, "/")
	if len(parts) <= 3 {
		return pkg
	}
	return "…/" + strings.Join(parts[len(parts)-2:], "/")
}

// termWidth returns the terminal column width using an ioctl syscall.
// Falls back to the COLUMNS env var, then 120.
type winsize struct{ Row, Col, Xpixel, Ypixel uint16 }

func termWidth() int {
	var ws winsize
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL,
		os.Stdout.Fd(), syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws))); e == 0 && ws.Col > 20 {
		return int(ws.Col)
	}
	if s := os.Getenv("COLUMNS"); s != "" {
		if w, err := strconv.Atoi(s); err == nil && w > 20 {
			return w
		}
	}
	return 120
}
