package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/didikizi/prism/internal/model"
	"github.com/didikizi/prism/internal/parser"
	"github.com/didikizi/prism/internal/render"
)

const version = "0.1.0"

func main() {
	noColor := flag.Bool("no-color", false, "disable colored output")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("prism v%s\n", version)
		os.Exit(0)
	}

	// Set before renderer is constructed so termenv picks it up.
	if *noColor {
		os.Setenv("NO_COLOR", "1")
	}

	state := model.New()
	r := render.New(state)
	r.Start()

	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer for large test output lines.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		ev, err := parser.ParseLine(scanner.Text())
		if err != nil {
			continue // skip non-JSON lines (build stderr, etc.)
		}
		r.HandleEvent(ev)
	}

	r.Finish()

	if state.HasFailures() {
		os.Exit(1)
	}
}
