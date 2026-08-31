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

- PR-only. Never self-merge a PR whose diff has content to judge --
  schema changes, generated output, description or page content,
  anything that changes what ships. Push a branch, open a PR, wait for
  the founder to review and merge.
- Self-merge is allowed only when a PR is purely mechanical: the
  identical file copied verbatim across repos (confirmed byte-for-byte
  before merging), a branch rebase or merge-in with no new content, or
  a version bump with no content change. A diff that mixes a
  mechanical change with anything else, even one line, needs review
  as a whole. When it is unclear which category a PR falls into, it
  needs review.
- Before pushing more commits to a branch with an open PR, confirm it is
  STILL open (`gh pr list --state open` or `gh pr view <n>`) — a merged PR's
  branch looks identical to any other from `git status` alone, and a push
  after merge lands nowhere near `main`, silently.
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

## Architecture documentation

A change to this runtime's own real contract — a new marker convention
(alongside `$computed`/`$secret`/`$ephemeral`), a new required field on
a generated binding's own runtime shape, a change to how a proposal's
resolved values reach the wire — is architectural and gets its
`ubiquex-internals` page (the developer documentation site) written or
updated in the same body of work, never a follow-up. A bug fix inside
the existing contract doesn't qualify. Matches `ubiquex` CLAUDE.md
rule 10.
