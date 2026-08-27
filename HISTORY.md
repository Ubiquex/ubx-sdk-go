# HISTORY.md — narrative archive

> Consulted only when a session needs to know why a decision was made, not on
> every open. For what's current, read `STATE.md` instead.

This file is new as of UBI-183 (2026-08-27). Real history predating it lives
in `ubiquex`'s own `HISTORY.md` (search `UBI-131`, `UBI-139`) and in this
repo's own real `git log`/merged-PR history, which is authoritative for what
actually shipped and when.

## Real, known decisions worth carrying forward

**UBI-131: a real "published" claim that wasn't.** A fix reported "committed
and pushed" across multiple session summaries meant only the monorepo's own
copy — this separate, real repo was never touched, still showing its
original scaffold commit a full day later. Caught only when the founder
pushed back and a real `git log` was run against this actual repo, not the
monorepo. `ubiquex`'s own CLAUDE.md rule 8 exists because of this incident —
its actual requirement: any session claiming a fix is "published" or "live"
for a shared runtime or per-provider bindings repo must verify against the
SEPARATE published repo/registry directly (a real `git log`/`diff` against
the actual separate repo, or a real registry query — the Go module proxy,
`jsr.io`, `pypi.org`), never infer "published" from a commit to the
monorepo's own copy alone.
