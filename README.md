# gojq

Pure-Go jq engine — the Krateo fork of [itchyny's gojq](https://github.com/itchyny/gojq),
carrying one behavioral patch: copy-on-write `del`/`delpaths` over caller-shared inputs.

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)
[![tag](https://img.shields.io/github/v/tag/krateo-platformops/gojq)](https://github.com/krateo-platformops/gojq/tags)

## What is this

A minimal fork of the upstream gojq library (fork point: upstream `v0.12.19`). The
whole delta is three files: an allocator-aware `deleteEmpty` so `del`/`delpaths` never
write into a shared input, a concurrent `-race` regression guard for that patch, and
the module identity. Everything else — the jq language, the CLI, the exported API —
is upstream, unmodified; this bundle documents only the delta and defers the rest.
Start at [docs/index.md](./docs/index.md).

## Install

Consumed as a Go module through a `replace` over the upstream path (code keeps
importing `github.com/itchyny/gojq`):

```
require github.com/itchyny/gojq v0.12.17

replace github.com/itchyny/gojq => github.com/krateo-platformops/gojq v0.13.0
```

See [docs/usage.md](./docs/usage.md) for why the replace-key shape is load-bearing.

## Configure

The library reads no configuration; behavior is set per-`Compile` via
`gojq.CompilerOption` values. Build tag `gojq_debug` (+ `GOJQ_DEBUG` env) enables the
debug tracer; the CLI honors `NO_COLOR` / `GOJQ_COLORS`. Details in
[docs/configuration.md](./docs/configuration.md).

## Examples

- [shared-input-cow](./examples/shared-input-cow/) — concurrent `del` over one shared
  input; race-clean and input-preserving on this fork (`go run -race .`).

## Docs

- [docs/index.md](./docs/index.md) — the map of this bundle
- [docs/overview.md](./docs/overview.md) — the fork delta vs upstream and why it exists
- [docs/usage.md](./docs/usage.md) — how Krateo components consume the fork
- [docs/configuration.md](./docs/configuration.md) — build tags / env / options surface
- [docs/api.md](./docs/api.md) — exported API (identical to upstream; deferred)
- [docs/examples.md](./docs/examples.md) — index of `examples/`
- [docs/release.md](./docs/release.md) — the actual tag-only release convention
- [docs/log.md](./docs/log.md) — curated fork history

## Develop & release

```sh
make build && make test   # go test -v -race ./... — includes the CoW fork guard
```

Releases are plain git tags (`vX.Y.Z`, continuing upstream's numbering) consumed via
the Go module `replace` pin — see [docs/release.md](./docs/release.md).
