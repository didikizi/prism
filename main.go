// Command prism turns `go test -json` output into a beautiful, streaming,
// screenshot-worthy report. Pipe usage:
//
//	go test -json ./... | prism
//
// Exit code mirrors the test result (non-zero on any failure), so it is safe
// to drop into CI pipelines.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/didikizi/prism/internal/gotest"
	"github.com/didikizi/prism/internal/ui"
)

const version = "0.1.0"

func main() {
	noColor := flag.Bool("no-color", false, "disable colored output")
	showVersion := flag.Bool("version", false, "print version and exit")
	bench := flag.String("bench", "both", "benchmark output: both | styled | md")
	flag.Parse()

	if *showVersion {
		fmt.Println("prism", version)
		return
	}
	if *noColor {
		os.Setenv("NO_COLOR", "1")
	}
	benchMode, err := ui.ParseBenchMode(*bench)
	if err != nil {
		fmt.Fprintln(os.Stderr, "prism:", err)
		os.Exit(2)
	}

	run := gotest.NewRun()
	live := ui.NewLive(os.Stdout, ui.IsTTY(os.Stdout))
	live.Start()

	// bufio.Reader (not Scanner) so a single huge output line never overflows
	// a fixed buffer and silently drops events.
	rd := bufio.NewReader(os.Stdin)
	for {
		line, err := rd.ReadString('\n')
		if s := strings.TrimRight(line, "\r\n"); s != "" {
			if ev, ok := gotest.Decode(s); ok {
				pkg, done := run.Add(ev)
				live.Counts(run.Pass, run.Fail, run.Skip)
				if done {
					live.PackageDone(pkg)
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "prism:", err)
			}
			break
		}
	}

	live.Stop()
	fmt.Fprint(os.Stdout, ui.Report(run, ui.Width(os.Stdout), benchMode))

	if run.Failed() {
		os.Exit(1)
	}
}
