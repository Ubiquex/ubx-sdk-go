# STATE.md — current state

> Rewritten, not appended, as the LAST act of every session. See `HISTORY.md`
> for the narrative.

## In flight

Nothing in flight as of 2026-08-27.

## Blocked

Nothing blocked. Zero open PRs.

## Current state

Latest known tag: `v0.1.2` (verified directly via `gh api
repos/Ubiquex/ubx-sdk-go/tags` — don't trust this file if it's gone stale,
re-check; also verify the Go module proxy directly, not just the tag).

## Before touching anything

- Never self-merge here. See `CLAUDE.md`.
- This is a SHARED runtime, not per-provider — a change here can ripple into
  every `ubx-sdk-<provider>` repo's own Go bindings AND `ubx` itself.
