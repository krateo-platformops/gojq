---
type: Architecture
title: gojq — the fork delta
description: What this fork changes vs upstream gojq v0.12.19 (allocator-aware copy-on-write deleteEmpty + module identity), why it exists, and the guard that keeps the patch alive.
resource: github.com/krateo-platformops/gojq
tags: [fork-delta, copy-on-write, allocator, snowplow]
timestamp: 2026-08-07T00:00:00Z
---

# The fork delta

This is a **minimal fork**. Fork point: upstream tag `v0.12.19` (the last upstream
commit in this repo's history is upstream's "bump up version to 0.12.19"). The full
diff versus that tag — verify with `git diff v0.12.19..main --stat` — is **three
files**:

| File | Change |
|---|---|
| `func.go` | allocator-aware `deleteEmpty` — the behavioral patch |
| `delete_empty_cow_test.go` | fork-only concurrent `-race` regression guard (new file) |
| `go.mod` | module identity `github.com/krateo-platformops/gojq` + in-fork `replace github.com/itchyny/gojq => ./` |

Everything else — parser, compiler, executor, builtins, CLI, docs — is upstream,
byte-identical. Do not look here for how gojq works; look at
[upstream](https://github.com/itchyny/gojq).

## Why the fork exists

snowplow (the Krateo portal content API) serves RESTAction results by running jq
`del`/`delpaths` filters over a **cached, shared** value: `api.listEnvelopeValue`
serves a *shallow envelope* that aliases the cached `entry.Items` instead of deep
copying it per request (the Ship 2a shallow-envelope optimization). That is only safe
if gojq treats its input as genuinely read-only.

Upstream gojq is *almost* there: the update path (`updateObject`,
`updateArrayIndex`, `updateArraySlice` in `func.go`) already copy-on-writes the
assignment spine. But the copy only covers the spine — sibling sub-trees of the
copied nodes stay **aliased to the caller's input** — and upstream's `deleteEmpty`
(the pass that removes the `struct{}{}` markers `delpaths` plants) recurses into
those aliased siblings and writes them **in place** (`v[k] = deleteEmpty(w)`,
`delete(v, k)`, in-slice compaction). Under concurrent serves of the same cached
value that is a data race, and on marker hits it is corruption of the shared input.

## What the patch does

`deleteEmpty(v, a)` now takes the allocator and uses `a.allocated(v)` to distinguish
node ownership:

- **gojq-owned node** (`a.allocated(v)` — created by gojq during this update pass):
  mutated in place, exactly like upstream. No extra allocation on the hot path.
- **non-owned node** (still aliased to the caller's input): survivors are copied
  into a freshly allocated node (`a.makeObject` / `a.makeArray`) and the recursion
  continues on the copy. The input is never written.

This mirrors the existing upstream CoW pattern in the update functions; `delpaths`
passes its allocator through (`deleteEmpty(u, a)`). Net effect: `del`/`delpaths`
(and everything built on them) are safe over a value shared with concurrent readers,
with allocation cost only on the aliased fringe actually visited.

## The guard: `delete_empty_cow_test.go`

The patch is protected by a fork-internal test designed to fail if a future
re-vendor of upstream drops it. Crucially it is **concurrent + `-race`**, not a
value compare: upstream's in-place write usually stores the *same* value back, so a
single-threaded before/after `reflect.DeepEqual` cannot see it — the write is only
observable as a data race when another goroutine reads the shared sub-tree, which is
exactly the production hazard. The guard runs `del`/`delpaths` over a shared input
from many goroutines under `-race` (`make test` runs `go test -v -race ./...`);
without the patch it reports dozens of races. It mirrors snowplow's in-process
`TestShip2a_DeleteEmpty_CoW_NoInputMutation`.

## Module identity (the replace-key shape)

The module path is `github.com/krateo-platformops/gojq`, but **all self-imports in
this repo say `github.com/itchyny/gojq`**, resolved by the in-fork
`replace github.com/itchyny/gojq => ./`. This is deliberate, not leftover:

- Consumers pin the fork with
  `replace github.com/itchyny/gojq => github.com/krateo-platformops/gojq vX.Y.Z`,
  so *they* know this module by the upstream path. If the fork's own test/cli files
  imported the fork path directly, a consumer's `go mod tidy` would see one module
  under two names ("used for two different module paths") and leave `go.sum`
  inconsistent.
- Standalone `go build`/`go test` of this repo resolves those `itchyny/gojq`
  imports back to `./` via the in-fork replace (which is ignored when the fork is
  consumed as a dependency — the consumer's own replace covers it there).

It also keeps the delta-vs-upstream diff minimal: no import rewrites, three files.

## Relationship to upstream

- Upstream remains authoritative for behavior, language and API; this fork tracks
  `v0.12.19` and intends re-vendors to be clean rebases of a 3-file delta.
- The patch is upstreamable in principle (it only extends upstream's own CoW
  invariant to `deleteEmpty`); until/unless it lands upstream, the `-race` guard is
  the contract that the delta survives every re-vendor.
