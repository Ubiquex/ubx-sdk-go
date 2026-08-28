# ubx-sdk-go

Shared Go runtime for `ubx`-generated per-provider SDK bindings.

A program built against this package never computes, never reaches a
provider, and never touches a ledger, it describes a desired end-state
(an `ubx:intent/v1` document) and stops. `resource()` returns a
`Computed[T]` reference, never a real value, so a resource's
not-yet-known attribute can still be wired into a sibling resource's
config at describe time.

## Where this sits

Not meant to be used directly. Every `ubx sdk gen --lang go` generated
bindings package (`ubx-sdk-aws`, `ubx-sdk-azure`, `ubx-sdk-google`,
`ubx-sdk-kubernetes`, `ubx-sdk-github`, `ubx-sdk-datadog`, one combined
repo per provider) imports `github.com/ubiquex/ubx-sdk-go/runtime` and
calls it against that binding's own generated `ResourceBinding`/Config
types. [ubiquex](https://github.com/ubiquex/ubiquex) is the real
source of truth this package's own hermetic evaluator (`goeval`) runs
programs against; it mounts this repo as a git submodule.

## What it contains

- `runtime/runtime.go`: the real, hand-maintained runtime, `Stack`,
  `Main`, `Resource`, `Computed[T]`, `CrossStack`, `Override`, and the
  blueprint-calling primitives
- `runtime/runtime_test.go`: hermetic tests

## How to use it

```
go get github.com/ubiquex/ubx-sdk-go/runtime
```

Not something you install standalone in practice, it arrives
transitively the moment a generated bindings package (`ubx-sdk-aws`
and the rest) is imported.

## How it's maintained

Hand-written and hand-maintained, not generated, unlike the
per-provider bindings packages that depend on it. Real engineering
sessions on this repo carry their own `STATE.md`/`HISTORY.md`/
`CLAUDE.md`, the same session-open protocol `ubiquex` itself uses.

## Links

- Docs: https://docs.ubiquex.io
- Internals (architecture and design): https://github.com/Ubiquex/ubiquex-internals
- Linear board: https://linear.app/ubiquex
