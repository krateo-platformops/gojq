---
type: Configuration
title: gojq — configuration
description: The whole (small) configuration surface — no config files; one build tag, two-and-a-half CLI env vars, and per-Compile options.
resource: github.com/krateo-platformops/gojq
tags: [build-tags, env, compiler-options]
timestamp: 2026-08-07T00:00:00Z
---

# Configuration

This is a library: it reads **no configuration files** and, in library form, **no
environment variables** (jq's `env`/`$ENV` are disabled unless the consumer opts in
via `gojq.WithEnvironLoader`). The fork adds no configuration of its own — the
copy-on-write `deleteEmpty` behavior is always on, not a knob. The full surface:

## Build tags

| Tag | Effect |
|---|---|
| `gojq_debug` | Compiles in the debug tracer (`debug.go`); at runtime it activates only when the `GOJQ_DEBUG` env var is set (any non-empty value; `GOJQ_DEBUG=stdout` selects stdout, anything else stderr). `make build-debug` builds the CLI with it; `make lint` staticchecks with it. |

## Environment variables (CLI only — `cmd/gojq`)

Read in `cli/cli.go`; irrelevant when embedding the library:

| Var | Effect |
|---|---|
| `NO_COLOR` (non-empty) or `TERM=dumb` | disables colored output |
| `GOJQ_COLORS` | custom color codes for the output (see upstream README for the format) |

## Per-Compile options (the library's real "configuration")

Behavior is configured programmatically per `gojq.Compile` call via
`CompilerOption` values — `WithModuleLoader`, `WithEnvironLoader`, `WithVariables`,
`WithFunction`, `WithIterFunction`, `WithInputIter`, … The surface is upstream's,
unchanged; see [api.md](./api.md) and the
[upstream reference](https://pkg.go.dev/github.com/itchyny/gojq#Compile).
