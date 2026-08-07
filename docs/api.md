---
type: API
title: gojq — exported Go API
description: The exported API surface is upstream gojq's, identical at v0.12.19 — the fork adds no exported identifiers; upstream's reference is authoritative.
resource: https://pkg.go.dev/github.com/itchyny/gojq
tags: [go-api, upstream-deferred]
timestamp: 2026-08-07T00:00:00Z
---

# API

The exported Go API of this fork is **byte-identical to upstream gojq at
`v0.12.19`**. The fork's delta is internal (`deleteEmpty` + one test + module
identity — see [overview](./overview.md)); it adds, removes and changes **no
exported identifiers**. The authoritative reference is upstream:

- **[pkg.go.dev — github.com/itchyny/gojq](https://pkg.go.dev/github.com/itchyny/gojq)**
  (browse at `v0.12.19` to match this fork's base exactly)

Remember the import path is `github.com/itchyny/gojq` even when consuming this fork
(the `replace`-key rule — [usage](./usage.md)).

## Entry points Krateo code actually uses

For orientation only (signatures and semantics: upstream reference above):

- `gojq.Parse(string) (*Query, error)` — parse a jq program.
- `gojq.Compile(*Query, ...CompilerOption) (*Code, error)` — compile once; a
  `*gojq.Code` is safe to reuse **read-only** across goroutines (snowplow reuses
  compiled codes across items and serves).
- `(*Code).Run(any, ...any) Iter` / `(*Code).RunWithContext` — execute; with this
  fork, `del`/`delpaths` inside the program never mutate the input value.
- `gojq.ModuleLoader` / `gojq.NewModuleLoader` — module loading (snowplow's
  `internal/support/jq` wraps this).
- `CompilerOption`s: `WithModuleLoader`, `WithEnvironLoader`, `WithVariables`,
  `WithFunction`, `WithIterFunction`, `WithInputIter`.

## The one semantic difference vs upstream

Not a signature change, a strengthened guarantee: **input values are genuinely
read-only under `del`/`delpaths`**, even for sub-trees upstream would have visited
in place after a copy-on-write of the assignment spine. Callers may hand `Run` a
value concurrently read (or served) elsewhere. This is the fork's entire reason to
exist; the `-race` guard `delete_empty_cow_test.go` pins it.
