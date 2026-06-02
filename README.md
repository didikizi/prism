# prism

**`go test` output that's actually worth looking at.**

[![CI](https://github.com/didikizi/prism/actions/workflows/ci.yml/badge.svg)](https://github.com/didikizi/prism/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/didikizi/prism.svg)](https://pkg.go.dev/github.com/didikizi/prism)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Install

```sh
go install github.com/didikizi/prism@latest
```

This drops the `prism` binary in `$(go env GOPATH)/bin`. Make sure that's on your
`PATH` (add `export PATH="$PATH:$(go env GOPATH)/bin"` to your shell profile if not).

## Try it

No install needed — clone and run the bundled demo (a fixture with failing,
panicking and racing tests) straight through prism:

```sh
git clone https://github.com/didikizi/prism && cd prism
make demo        # assertions + subtests + a panic
make demo-race   # the same, with the -race data-race card
```

## Usage

Pipe any `go test -json` run through `prism`:

```sh
go test -json ./... | prism            # whole module
go test -json . | prism                # current package
go test -race -json ./... | prism      # surface data races as RACE cards
```

Handy shell alias:

```sh
alias gt='go test -json ./... | prism'
```

Works in CI too — the exit code mirrors the test result (non-zero on any failure),
so it slots straight into a pipeline. When stdout is not a TTY (a pipe or log
redirect) the spinner is suppressed automatically and colour is dropped, keeping
logs clean.

## VS Code

prism is a CLI, not an extension — you run it from VS Code's **integrated
terminal**, which is a real TTY, so the spinner, colours and cards all render:

```sh
go test -json ./... | prism
```

### Auto-apply to every test and benchmark run

Wire prism into VS Code with two tasks — one for tests, one for benchmarks —
in `.vscode/tasks.json`. Commit this file and it applies to everyone on the repo:

```jsonc
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "test: prism",
      "type": "shell",
      "command": "go test -json ./... | prism",
      "group": { "kind": "test", "isDefault": true },
      "presentation": { "clear": true, "reveal": "always", "panel": "dedicated" },
      "problemMatcher": []
    },
    {
      "label": "bench: prism",
      "type": "shell",
      "command": "go test -run=^$ -bench=. -benchmem -json ./... | prism",
      "group": "test",
      "presentation": { "clear": true, "reveal": "always", "panel": "dedicated" },
      "problemMatcher": []
    }
  ]
}
```

Run the test task with **Tasks: Run Test Task** (`⌘⇧P` / `Ctrl⇧P` → that command,
or bind a key to `workbench.action.tasks.test`). For benchmarks, run
**Tasks: Run Task → bench: prism**.

**Bind both to keys** in `keybindings.json` so it's truly one keystroke:

```jsonc
{ "key": "cmd+shift+t", "command": "workbench.action.tasks.test" },
{ "key": "cmd+shift+b", "command": "workbench.action.tasks.runTask", "args": "bench: prism" }
```

**Run automatically on folder open** — add this to the *test* task to have prism
greet you with a full run every time the project opens:

```jsonc
"runOptions": { "runOn": "folderOpen" }
```

> Note: `-run=^$` in the bench task skips tests so the benchmark phase always runs
> (`go test` skips benchmarks when tests fail). These tasks run *alongside* the Go
> extension's built-in test runner — they don't replace it. The extension captures
> its own output, so prism reads from the terminal pipe, not the extension's panel.

## What you get

- **Live spinner** with running pass/fail/skip counters while tests execute
- **Per-package result lines** as each package completes
- **Fail cards** — each failure in its own rounded box: status badge, test name,
  `package · file:line`, and the assertion message (de-indented, noise stripped)
- **Panic detection** — `PANIC` card with the panic headline up top and a dimmed,
  truncated stack trace below
- **Race detection** — run with `-race`; data races render as a distinct `RACE` card
- **Build error cards** — compiler errors shown inline, not swallowed
- **Benchmark table** — a styled panel with relative-speed bars **plus** a
  copy-ready Markdown table you can paste straight into a README or PR
- **Final summary panel** — totals + top-5 slowest tests, bordered green or red

## Flags

| Flag | Effect |
|------|--------|
| `--bench <mode>` | Benchmark output: `both` (default), `styled` (screenshot), `md` (copy) |
| `--no-color` | Disable all color (plain text, good for log files) |
| `--version` | Print version and exit |

## Benchmarks

Pipe a `-bench` run through prism and you get a styled panel with relative-speed
bars **and** a Markdown table ready to paste into your project:

```sh
go test -bench=. -benchmem -json ./... | prism
```

```
╭────────────────────────────────────────────────────────────────────────────╮
│  benchmarks                                                                │
│  linux/amd64 · Intel(R) Xeon(R) CPU E5-2697 v3 @ 2.60GHz                   │
│                                                                            │
│    Add                       0.49 ns   ▏                  0 B   0 allocs  │
│    Itoa                     43.87 ns   ▏                  7 B   0 allocs  │
│    SortSmall               203.00 ns   ▎                104 B   2 allocs  │
│    AllocLarge               12.74 µs   ████████████   32.0 kB   2 allocs   │
╰────────────────────────────────────────────────────────────────────────────╯
```

…and below it, ready to paste into a README or PR (the hardware line carries over
as an italic caption):

_linux/amd64 · Intel(R) Xeon(R) CPU E5-2697 v3 @ 2.60GHz_

| Benchmark           |  ns/op |  B/op | allocs/op |
| ------------------- | -----: | ----: | --------: |
| BenchmarkAdd        | 0.4925 |     0 |         0 |
| BenchmarkItoa       |  43.87 |     7 |         0 |
| BenchmarkSortSmall  |    203 |   104 |         2 |
| BenchmarkAllocLarge |  12742 | 32792 |         2 |

The bar is each benchmark's time relative to the slowest, and the header records
the machine the run was on. Pick the format with `--bench`: `styled` for a clean
screenshot, `md` to redirect a table into a file
(`go test -bench=. -json ./... | prism --bench=md >> BENCHMARKS.md`).

---

> **weekend project · v0.1 · experimental**  
> Feedback and issues welcome — open a [GitHub issue](https://github.com/didikizi/prism/issues) or ping [@didikizi](https://github.com/didikizi).
