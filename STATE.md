# STATE.md — current state

> Rewritten, not appended, as the LAST act of every session. See `HISTORY.md`
> for the narrative.

## In flight

Nothing in flight as of 2026-09-13.

## Blocked

Nothing blocked.

## Current state

Latest tag and release: **`v0.6.0`** (2026-09-13), at commit `374c45b`.

Verified three ways rather than from the tag alone, since a tag proves
nothing about what a consumer can actually fetch:

- `go list -m -versions github.com/ubiquex/ubx-sdk-go` reports it
- `proxy.golang.org/.../@v/v0.6.0.info` resolves it to `374c45b`
- a real external program, no `replace` directive, fetched it from the
  proxy and ran under `ubx`, producing the attributed provenance
  `v0.6.0` adds. The same program against `v0.5.0` produced none, which
  is what proves the change is in the release rather than somewhere else.

`v0.6.0` carries UBI-266, call-site attribution: a blueprint written as
code now gets its resources attributed to it, where before they carried
no provenance at all. Nothing else changed since `v0.5.0`.

## Before touching anything

- Never self-merge here. See `CLAUDE.md`.
- This is a SHARED runtime, not per-provider — a change here can ripple into
  every `ubx-sdk-<provider>` repo's own Go bindings AND `ubx` itself.
- This file said `v0.1.2` for four releases. Re-check with `gh api
  repos/Ubiquex/ubx-sdk-go/tags` and the module proxy rather than
  trusting it, and update it when you cut a release.
