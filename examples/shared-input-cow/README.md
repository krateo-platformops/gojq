---
type: Example
title: gojq — shared-input copy-on-write example
description: Concurrent del/delpaths over one shared input — race-clean and input-preserving on this fork, a data race on upstream; run with go run -race.
resource: github.com/krateo-platformops/gojq
tags: [copy-on-write, race, delpaths]
timestamp: 2026-08-07T00:00:00Z
---

# shared-input-cow

Demonstrates the fork's entire behavioral delta ([overview](../../docs/overview.md)):
8 goroutines run `delpaths` over **one shared input** while 4 readers walk it, then
the input is compared against a snapshot.

**Preconditions:** a checkout of this repo (the example imports the replace-key path
`github.com/itchyny/gojq`, resolved to the checkout by the in-fork `replace`) and
Go ≥ 1.24. Run **with `-race`** — the in-place write this fork removes is usually a
same-value store, invisible to a value compare and only observable as a race:

```sh
go run -race ./examples/shared-input-cow
```

Expected output (this fork): `ok: 1600 concurrent delpaths runs, shared input untouched: …`.
On upstream gojq `v0.12.19` the same program fails under `-race` with data races in
`deleteEmpty`.
