# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-06-03

### Fixed
- Build failures now render the compiler error: `build-output`/`build-fail`
  events (keyed by `ImportPath`, linked via `FailedBuild`) are parsed and
  attached to the package, instead of being dropped into a phantom package

### Changed
- A parent test whose only failure is a failing subtest is no longer shown as an
  empty card — the leaf `Parent/Child` card carries the detail

### Added
- Golden tests: `internal/ui/testdata/*.jsonl` fixtures → `*.golden` expected
  output, covering pass / assertion / panic / race / build / subtests / benchmarks
  (regenerate with `go test ./internal/ui -run TestReportGolden -update`)

## [0.1.0] - 2026-05-30

### Added
- Streaming pipe-filter: `go test -json ./... | prism`
- Live spinner with pass / fail / skip counters while tests run (suppressed when stdout is not a TTY)
- Per-package result lines printed as each package completes
- Fail cards — status badge, test name, `package · file:line`, and the cleaned assertion message
- Panic detection — `PANIC` card with the panic headline and a dimmed, truncated stack trace
- Race detection — `-race` data races rendered as a distinct `RACE` card
- Build error cards — compiler failures shown inline instead of being swallowed
- Benchmark support — `-bench` results parsed and rendered as a styled panel with
  relative-speed bars plus a copy-ready Markdown table; `--bench both|styled|md`.
  Captures the run hardware (goos/goarch/cpu) and shows it under the panel and as
  an italic caption above the Markdown table
- Final summary panel with totals and an aligned top-5 slowest-tests table
- Border colour reflects overall result (green = all pass, red = any failure)
- Exit code mirrors test result: non-zero on any failure (CI-safe)
- `--no-color` flag for plain-text output
- `--version` flag
- Colour palette based on [Catppuccin Mocha](https://github.com/catppuccin/catppuccin)

[Unreleased]: https://github.com/didikizi/prism/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/didikizi/prism/releases/tag/v0.1.0
