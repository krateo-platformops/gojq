---
type: Library
title: gojq — index
description: The map of the gojq fork doc bundle — Krateo's minimal fork of the upstream gojq jq engine, patched for copy-on-write del/delpaths over shared inputs.
resource: github.com/krateo-platformops/gojq
tags: [library, fork, jq, gojq]
timestamp: 2026-08-07T00:00:00Z
---

# gojq (Krateo fork)

This repo is a **fork of [itchyny's gojq](https://github.com/itchyny/gojq)** — the
pure-Go jq implementation — kept as close to upstream as possible. Fork point:
upstream **`v0.12.19`**. The entire delta versus upstream is three files
(`func.go`, `delete_empty_cow_test.go`, `go.mod`): `del`/`delpaths` are made
copy-on-write over any node the caller shares with gojq, so a consumer can hand the
engine a **shared, read-only input** without a per-call deep copy. snowplow's
RESTAction refilter path is the consumer that needs this.

This bundle documents **the fork delta and how Krateo consumes it** — it does not
duplicate upstream's documentation. For the jq language, the CLI and the full library
API, upstream is authoritative: [upstream repo](https://github.com/itchyny/gojq),
[pkg.go.dev](https://pkg.go.dev/github.com/itchyny/gojq).

## The bundle (start here)

- [overview](./overview.md) — the fork delta vs upstream `v0.12.19`, why it exists,
  and the module-identity mechanics (replace-key self-imports).
- [usage](./usage.md) — how Krateo components consume the fork: the `replace`-pin
  go.mod shape, verbatim from snowplow; standalone build/test.
- [configuration](./configuration.md) — the whole (small) config surface: the
  `gojq_debug` build tag, the env vars the CLI reads, compile-time options.
- [api](./api.md) — the exported Go API: identical to upstream; the fork adds no
  exported identifiers. Deferred to upstream's reference.
- [examples](./examples.md) — the runnable example under `examples/`.
- [release](./release.md) — the actual release convention: tag-only Go library,
  `vX.Y.Z` continuing upstream's numbering, consumed via the `replace` pin.
- [log](./log.md) — curated fork history (upstream history stays in
  [CHANGELOG.md](../CHANGELOG.md)).
- [llms.txt](./llms.txt) — the version-pinned agent index of this bundle.
