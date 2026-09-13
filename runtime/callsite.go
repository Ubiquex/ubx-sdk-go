package sdk

// callsite.go attributes a resource to the blueprint whose code created
// it, by looking at the call stack rather than at a marker the
// blueprint had to push for itself.
//
// The problem it solves: PushBlueprintSource is called only by
// generated code, which an Ubxfile blueprint's build produces and a
// blueprint written as code does not have. A code blueprint is an
// ordinary hand-written function, and the as-code model's whole point
// is that nothing is declared twice, so nothing marks its resources and
// every one of them reached the ledger with no source at all.
//
// How the roots get here: `ubx` discovers, before running this program,
// every blueprint whose code it can reach, and links the answer into
// this variable with `go build -ldflags -X`. A linker flag rather than
// an environment variable because the evaluator runs this program with
// an empty environment on purpose, and reading a file would mean
// reaching outside the sandbox. Nothing here ever computes a content
// hash: this side reports a bare NAME, and the host completes it
// afterwards from the same discovery pass that produced the roots.
//
// A program built by anything other than `ubx` has this unset, so every
// check below short-circuits on an empty list and an ordinary stack
// pays one nil check per resource.

import (
	"encoding/base64"
	"encoding/json"
	"runtime"
	"strings"
	"sync"
)

// blueprintRootsB64 is set at link time by ubx's own Go evaluator
// (`-X github.com/ubiquex/ubx-sdk-go/runtime.blueprintRootsB64=...`).
//
// Base64 of a JSON array, because `go build` splits an -ldflags value
// on spaces itself and quoting rules differ by platform. Never written
// at runtime, and deliberately unexported: this is a channel between
// ubx and this runtime, not API.
var blueprintRootsB64 string

// blueprintRoot is one blueprint this program can reach. Match is a
// package import-path prefix, compared against a frame's own
// fully-qualified function name. It is an import path rather than a
// directory because the file paths compiled into a binary are the build
// machine's, and this process has no reason to believe they mean
// anything where it runs.
type blueprintRoot struct {
	Match string `json:"match"`
	Name  string `json:"name"`
}

// parseBlueprintRoots decodes the linked manifest once. A malformed or
// undecodable value yields no roots rather than a panic: this mechanism
// only ever ADDS provenance, so failing to read it must degrade to the
// behaviour that existed before it, never take down an evaluation.
var parseBlueprintRoots = sync.OnceValue(func() []blueprintRoot {
	return decodeBlueprintRoots(blueprintRootsB64)
})

func decodeBlueprintRoots(encoded string) []blueprintRoot {
	if encoded == "" {
		return nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}
	var roots []blueprintRoot
	if err := json.Unmarshal(raw, &roots); err != nil {
		return nil
	}
	return roots
}

// callSiteBlueprint returns the name of the blueprint whose code is
// innermost on the current call stack, or "" when the call did not come
// from inside any known blueprint.
//
// Innermost, not outermost: a blueprint calling another blueprint's
// function produces a stack with both on it, and the resource belongs
// to whichever one actually called Resource().
func callSiteBlueprint() string {
	roots := parseBlueprintRoots()
	if len(roots) == 0 {
		return ""
	}

	// Skipping 2 leaves out Callers and callSiteBlueprint itself. The
	// rest of this runtime's own frames are skipped by the matching
	// instead of by a count, since the number of frames between
	// Resource() and here is an implementation detail that would
	// otherwise silently become load-bearing.
	pcs := make([]uintptr, 64)
	n := runtime.Callers(2, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if name := blueprintForFunction(frame.Function, roots); name != "" {
			return name
		}
		if !more {
			break
		}
	}
	return ""
}

// blueprintForFunction reports which root, if any, owns a frame's
// fully-qualified function name (e.g.
// "github.com/ubx-blueprints/widget-bp.BuildWidget").
//
// The boundary character matters. A bare strings.HasPrefix would let
// module "example.com/bp" claim a function in "example.com/bpother",
// which is a different module that merely starts the same way. A real
// frame name continues with "." for a function in the module's own root
// package, or "/" for one in a subpackage, and with nothing else.
func blueprintForFunction(function string, roots []blueprintRoot) string {
	if function == "" {
		return ""
	}
	for _, r := range roots {
		if r.Match == "" || r.Name == "" {
			continue
		}
		if !strings.HasPrefix(function, r.Match) {
			continue
		}
		rest := function[len(r.Match):]
		if strings.HasPrefix(rest, ".") || strings.HasPrefix(rest, "/") {
			return r.Name
		}
	}
	return ""
}
