# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-30

### Added
- Streaming pipe-filter: `go test -json ./... | prism`
- Live spinner with pass / fail / skip counters while tests run
- Per-package result lines printed as each package completes
- Fail cards — each failure in its own rounded box with test name, package, and assertion output
- Panic detection — raw panics rendered as a distinct PANIC card with full goroutine trace
- Final summary panel with totals and top-5 slowest tests
- Border colour reflects overall result (green = all pass, red = any failure)
- Exit code mirrors test result: non-zero on any failure (CI-safe)
- `--no-color` flag for plain-text output
- `--version` flag
- Colour palette based on [Catppuccin Mocha](https://github.com/catppuccin/catppuccin)

[Unreleased]: https://github.com/didikizi/prism/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/didikizi/prism/releases/tag/v0.1.0
