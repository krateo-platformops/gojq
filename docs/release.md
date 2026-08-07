---
type: Runbook
title: gojq — release
description: The actual release convention — a tag-only Go library; vX.Y.Z tags continuing upstream's numbering, consumed via the go.mod replace pin; no OCI, no images, no GitHub Releases.
resource: github.com/krateo-platformops/gojq
tags: [release, tags, go-modules]
timestamp: 2026-08-07T00:00:00Z
---

# Release

This is a **tag-only Go library**. A release is a git tag; consumption is the Go
module proxy resolving that tag through the consumer's `replace` pin. There is no
OCI chart, no published container image, and — as of `v0.13.0` — no GitHub Release
objects and no recorded Actions runs on this repo (verified against the live repo:
zero releases, zero workflow runs).

## Tag convention (derived from the existing tags)

- **`vX.Y.Z`, `v`-prefixed** — this fork keeps upstream's tag style (unlike Krateo
  component monorepos, which tag without the prefix). Fork tags continue upstream's
  numbering line: upstream fork point `v0.12.19` → fork `v0.12.20`, `v0.12.21` →
  **`v0.13.0`** (current; the first tag under the `github.com/krateo-platformops/gojq`
  module identity — the only line consumers can pin, see [log](./log.md) for what
  the earlier fork tags were).
- The minor bump to `0.13.0` marks the module-identity migration, leaving the
  `0.12.x` space to upstream's own line.

## Runbook

1. **Merge to `main`** — keep the delta minimal (3 files vs upstream; see
   [overview](./overview.md)).
2. **Verify locally** (there is no CI gate to lean on — zero recorded runs):

   ```sh
   make test   # go test -v -race ./... — the CoW guard MUST run under -race
   make lint
   ```

3. **Tag and push**:

   ```sh
   git tag v0.13.1 && git push origin v0.13.1
   ```

4. **Bump the consumer pin** — edit the `replace` line in each consumer
   (today: snowplow's `go/snowplow/go.mod`) to the new version and `go mod tidy`:

   ```
   replace github.com/itchyny/gojq => github.com/krateo-platformops/gojq v0.13.1
   ```

## What `ci.yaml` says vs what actually happens

[.github/workflows/ci.yaml](../.github/workflows/ci.yaml) is inherited from
upstream: on `v*` tags it would cross-build the CLI, push Docker images and create a
GitHub Release with binaries. **None of that has ever run on this repo** (no
workflow runs recorded; no packages published). The library release path does not
need it — the git tag alone is the artifact. Treat the CLI/Docker/Release jobs as
upstream inheritance, not as this fork's release process. The `lint-docs` job
(appended by the docs-standard adoption) is the part of `ci.yaml` this repo relies
on for PRs.

## Re-vendoring upstream

A new upstream version = rebase the 3-file delta onto the upstream tag, keep the
fork's `go.mod` identity block, run `make test` (the `-race` guard fails loudly if
the `deleteEmpty` patch is dropped), then tag per the convention above.
