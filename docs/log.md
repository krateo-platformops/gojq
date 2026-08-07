---
type: Log
title: gojq — curated fork history
description: The fork's notable events, from the upstream v0.12.19 fork point to the krateo-platformops module identity; upstream history stays in CHANGELOG.md.
resource: github.com/krateo-platformops/gojq
tags: [history, fork]
timestamp: 2026-08-07T00:00:00Z
---

# Log

Curated fork history, newest first. Upstream's own history (through the fork point)
is [CHANGELOG.md](../CHANGELOG.md) and upstream's releases — not duplicated here.

- **2026-08-07 — docs standard adoption.** This bundle; `lint-docs` appended to the
  PR CI.
- **2026-08-03 — `v0.13.0`: module identity migrated to `krateo-platformops`**
  (commit `823fdc6`), part of the org-wide move to full independence. One-line
  change (the `module` path); self-imports already used the replace-key shape.
  snowplow pins this tag. Earlier fork tags carry the interim module path and are
  historical only.
- **2026-05-29 — `v0.12.21`: replace-key self-imports** (commit `19ddd8a`). All
  self-imports switched to the upstream path `github.com/itchyny/gojq` + the in-fork
  `replace … => ./`, so a consumer's versioned `replace` no longer sees the module
  under two names (`go mod tidy`/`go.sum` breakage). See
  [overview](./overview.md#module-identity-the-replace-key-shape).
- **2026-05-29 — `v0.12.20` (also tagged `v0.12.19-snowplow.cow.1`): the fork's
  reason to exist** (commit `35c7261`). Allocator-aware `deleteEmpty`:
  `del`/`delpaths` copy-on-write every node not owned by gojq's allocator instead of
  mutating the caller's shared input in place, unlocking snowplow's Ship 2a
  shallow-envelope serve (no per-request deep copy of cached items). Ships with the
  concurrent `-race` CoW guard (`delete_empty_cow_test.go`) that fails if a future
  re-vendor drops the patch. The module path at this point was an interim
  personal-org path, superseded at `v0.13.0`.
- **2026-04-01 — fork point: upstream `v0.12.19`.** Last shared commit with
  upstream (`b7ebffb`, upstream's version bump). Everything before this is upstream
  history.
