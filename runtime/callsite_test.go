package sdk

import "testing"

// The matching rule, which is where this can go quietly wrong: a bare
// prefix check would let one module claim another that merely starts
// the same way.
func TestBlueprintForFunction_BoundaryIsRespected(t *testing.T) {
	roots := []blueprintRoot{
		{Match: "github.com/ubx-blueprints/widget-bp", Name: "widget-bp"},
		{Match: "example.com/bp", Name: "bp"},
	}

	for _, c := range []struct {
		function string
		want     string
	}{
		{"github.com/ubx-blueprints/widget-bp.BuildWidget", "widget-bp"},
		{"github.com/ubx-blueprints/widget-bp/internal.helper", "widget-bp"},
		{"github.com/ubx-blueprints/widget-bp.(*T).Method", "widget-bp"},
		{"example.com/bp.Build", "bp"},

		// A different module that merely starts the same way.
		{"github.com/ubx-blueprints/widget-bp-other.Build", ""},
		{"example.com/bpother.Build", ""},

		{"main.main", ""},
		{"github.com/ubiquex/ubx-sdk-go/runtime.Resource", ""},
		{"", ""},
	} {
		if got := blueprintForFunction(c.function, roots); got != c.want {
			t.Errorf("blueprintForFunction(%q) = %q, want %q", c.function, got, c.want)
		}
	}
}

// An unlinked program, which is every program not built by ubx, has to
// behave exactly as it did before this mechanism existed.
func TestCallSiteBlueprint_UnlinkedIsInert(t *testing.T) {
	if got := callSiteBlueprint(); got != "" {
		t.Fatalf("callSiteBlueprint() = %q with no roots linked, want empty", got)
	}
}

// A malformed manifest degrades to no roots. This mechanism only ever
// adds provenance, so failing to read it must never take down an
// evaluation.
func TestDecodeBlueprintRoots_MalformedYieldsNothing(t *testing.T) {
	for _, bad := range []string{"not base64 !!!", "bm90IGpzb24=", ""} {
		if roots := decodeBlueprintRoots(bad); len(roots) != 0 {
			t.Errorf("decodeBlueprintRoots(%q) = %v, want none", bad, roots)
		}
	}
}

// The exact bytes blueprint.EncodeBlueprintRootManifest produces, so
// the two halves of this contract are pinned against each other rather
// than each against its own idea of the format.
func TestDecodeBlueprintRoots_RoundTrip(t *testing.T) {
	const encoded = "W3sibWF0Y2giOiJleGFtcGxlLmNvbS9icCIsIm5hbWUiOiJicCJ9XQ=="
	roots := decodeBlueprintRoots(encoded)
	if len(roots) != 1 || roots[0].Match != "example.com/bp" || roots[0].Name != "bp" {
		t.Fatalf("decodeBlueprintRoots = %+v", roots)
	}
}
