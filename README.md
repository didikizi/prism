# prism

**`go test` output that's actually worth looking at.**

<!-- TODO: replace with demo.gif once recorded with `make record` -->
> *GIF coming soon — run `make demo` to see it live*

---

## Install

```sh
go install github.com/didikizi/prism@latest
```

## Usage

```sh
go test -json ./... | prism
```

Works in CI too — exit code mirrors the test result (non-zero on any failure).

## What you get

- **Live spinner** with running pass/fail/skip counters while tests execute
- **Per-package result lines** as each package completes
- **Fail cards** — each failure in its own rounded box: test name, package, assertion output
- **Panic detection** — raw panics rendered as a distinct PANIC card with full trace
- **Final summary panel** — totals + top-5 slowest tests, bordered in green or red

## Flags

| Flag | Effect |
|------|--------|
| `--no-color` | Disable all color (plain text, good for log files) |
| `--version` | Print version and exit |

---

