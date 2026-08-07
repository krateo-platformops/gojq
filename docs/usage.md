---
type: Usage
title: gojq — usage
description: How Krateo components consume this fork — the require+replace go.mod shape (verbatim from snowplow), the import path rule, and standalone build/test.
resource: github.com/krateo-platformops/gojq
tags: [go-modules, replace, snowplow]
timestamp: 2026-08-07T00:00:00Z
---

# Usage

## Consuming the fork (the only supported shape)

Krateo components consume this fork as a Go module **through a `replace` over the
upstream path**. Verbatim from snowplow's `go/snowplow/go.mod` (the one current
consumer):

```
require github.com/itchyny/gojq v0.12.17

replace github.com/itchyny/gojq => github.com/krateo-platformops/gojq v0.13.0
```

Code then imports the **upstream path** and gets the fork:

```go
import "github.com/itchyny/gojq"

query, err := gojq.Parse(`del(.paths.to.remove)`)
```

Two rules, both load-bearing (see [overview](./overview.md) for the mechanics):

1. **Never import `github.com/krateo-platformops/gojq` directly.** The import path
   is always `github.com/itchyny/gojq`; only the `replace` right-hand side names the
   fork. Importing the fork path alongside the replace makes `go mod tidy` see one
   module under two names and breaks `go.sum` (`go build -mod=readonly` fails).
2. **Bump the pin by editing the `replace` version** (currently `v0.13.0` — the
   first tag under the `krateo-platformops` module identity; earlier fork tags carry
   a superseded module path and do not resolve for this replace).

What the fork buys the consumer: `del`/`delpaths` are copy-on-write over any node
gojq does not own, so the input value can be **shared** (e.g. snowplow's cached
`entry.Items` served through a shallow envelope) without a per-call deep copy —
concurrent serves over the same cached value are race-clean.

## Where snowplow uses it

- `go/snowplow/internal/resolvers/restactions/api/` — `refilter.go` / `jqvalue.go`:
  compiled `*gojq.Code` reused read-only across items; `del`/`delpaths` refiltering
  over the shared cached value (the reason this fork exists).
- `go/snowplow/internal/support/jq/modules.go` — the shared `gojq.ModuleLoader`.

## Standalone build / test (working on the fork itself)

```sh
git clone https://github.com/krateo-platformops/gojq.git
cd gojq
make build   # builds the (upstream, unmodified) CLI: ./gojq
make test    # go test -v -race ./... — MUST stay -race: it powers the CoW guard
make lint    # go vet + staticcheck
```

Standalone builds resolve the `itchyny/gojq` self-imports to the checkout via the
in-fork `replace github.com/itchyny/gojq => ./` in [go.mod](../go.mod).

For jq-language usage, CLI flags and general library patterns, defer to
[upstream's README](https://github.com/itchyny/gojq#readme) — they apply unchanged.
