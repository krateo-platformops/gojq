---
type: ExampleIndex
title: gojq — examples
description: Index of the runnable examples under examples/ — one line each.
resource: github.com/krateo-platformops/gojq
tags: [examples]
timestamp: 2026-08-07T00:00:00Z
---

# Examples

- [shared-input-cow](../examples/shared-input-cow/) — the fork's delta, demonstrated:
  many goroutines run `del`/`delpaths` over ONE shared input while readers walk it;
  run with `go run -race .` — race-clean and input-preserving on this fork, a data
  race on upstream. Compilable Go `main` package; builds with the repo
  (`go build ./...`).

The upstream `_gojq/` and `_tools/` directories are upstream's own (underscore =
outside the build); they are not examples of this fork's delta.
