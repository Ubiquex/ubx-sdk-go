# CLAUDE.md — ubx-sdk-go

## What this is

The shared Go runtime every `ubx-sdk-<provider>`'s own generated `sdk/go/`
bindings depend on — the real, hand-written (not generated) library code:
proposal helpers, blueprint interfaces, whatever every per-provider Go
binding needs in common. Coordinating repo: `github.com/ubiquex/ubiquex`
(this is also `ubx`'s own `go:embed` build input, per this repo's own real
description — a change here can affect `ubx` itself, not just downstream
SDK consumers).

## Session protocol

1. Read `STATE.md` first — current state only, rewritten not appended.
2. `STATE.md` is rewritten, not appended, as the LAST act of every session.
   Anything that becomes history moves to `HISTORY.md`.
3. Only reference Linear issue IDs given in the handoff prompt; never infer
   one.

## Git rules (strict)

- PR-only. Never self-merge — push a branch, open a PR, wait for the founder.
- NO AI attribution anywhere in commits or PR bodies.

## Publishing discipline

- This is a shared runtime, not a per-provider repo — a bug here can affect
  every `ubx-sdk-<provider>` repo AND `ubx` itself (`go:embed` build input).
  Verify the real, separate published module directly (Go module proxy or
  `gh api repos/Ubiquex/ubx-sdk-go/tags`) before claiming a fix is live — a
  commit to this repo's own `main` is NOT the same as "published". Never
  infer "published" from a commit to the monorepo's own copy alone
  (`ubiquex`'s own CLAUDE.md rule 8). This exact class of mistake already
  happened to THIS repo once (UBI-131): a Go fix was reported "committed and
  pushed" across multiple session summaries, but only the monorepo's own
  copy had changed — this repo itself was never touched, still showing its
  original scaffold commit a full day later, caught only when the founder
  pushed back on the status claim and a real `git log` was run against this
  actual repo, not the monorepo.
