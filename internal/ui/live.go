package ui

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/didikizi/prism/internal/gotest"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Live renders the streaming phase: a transient spinner line that tracks live
// counters, with permanent per-package result lines printed above it.
//
// On a TTY it animates via "\r\033[K" (carriage return + clear-to-end-of-line).
// When stdout is not a TTY the spinner is suppressed entirely so pipes and CI
// logs stay clean; package lines are still printed.
type Live struct {
	w   io.Writer
	tty bool

	mu               sync.Mutex
	frame            int
	pass, fail, skip int
	active           bool // a transient line is currently on screen

	stop chan struct{}
	done chan struct{}
}

func NewLive(w io.Writer, tty bool) *Live {
	return &Live{w: w, tty: tty}
}

// Start launches the animation loop (no-op when not a TTY).
func (l *Live) Start() {
	if !l.tty {
		return
	}
	l.stop = make(chan struct{})
	l.done = make(chan struct{})
	go func() {
		defer close(l.done)
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-l.stop:
				return
			case <-t.C:
				l.mu.Lock()
				l.frame = (l.frame + 1) % len(spinnerFrames)
				l.draw()
				l.mu.Unlock()
			}
		}
	}()
}

// Counts updates the running totals shown by the spinner.
func (l *Live) Counts(pass, fail, skip int) {
	l.mu.Lock()
	l.pass, l.fail, l.skip = pass, fail, skip
	l.mu.Unlock()
}

// PackageDone prints a permanent result line for a finished package.
func (l *Live) PackageDone(p *gotest.Package) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clear()
	fmt.Fprintln(l.w, packageLine(p))
}

// Stop ends the animation and clears the transient line.
func (l *Live) Stop() {
	if l.tty {
		close(l.stop)
		<-l.done
	}
	l.mu.Lock()
	l.clear()
	l.mu.Unlock()
}

// clear erases the transient line if one is present. Caller holds the lock.
func (l *Live) clear() {
	if l.active {
		fmt.Fprint(l.w, "\r\033[K")
		l.active = false
	}
}

// draw paints the spinner line. Caller holds the lock.
func (l *Live) draw() {
	l.clear()
	line := stSpin.Render(spinnerFrames[l.frame]) + "  " +
		stDim.Render("running") + "    " +
		stPass.Render(fmt.Sprintf("%d %s", l.pass, glyphPass)) + "   " +
		stFail.Render(fmt.Sprintf("%d %s", l.fail, glyphFail)) + "   " +
		stSkip.Render(fmt.Sprintf("%d %s", l.skip, glyphSkip))
	fmt.Fprint(l.w, line)
	l.active = true
}
