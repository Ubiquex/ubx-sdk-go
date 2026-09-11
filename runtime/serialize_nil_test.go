package sdk

import (
	"encoding/json"
	"testing"
)

// An unset optional field must be OMITTED, not sent as null.
//
// A hand-written blueprint Config says "optional" with a pointer field
// and "not given" with nil. Before this, a nil pointer serialized as an
// explicit null, and that difference is not cosmetic: SQS rejects
// FifoQueue outright on a standard queue rather than defaulting it, so
// an attribute the author meant to omit came back as a 400 from the
// provider (UBI-258).

type nilProbeConfig struct {
	Required string
	Optional *int
	AlsoSet  *bool
	Opaque   any
}

func TestSerializeConfig_NilPointerIsOmittedNotNull(t *testing.T) {
	binding := ResourceBinding{
		WireType: "probe_thing",
		Fields: FieldMap{
			"Required": FieldSpec{WireName: "required"},
			"Optional": FieldSpec{WireName: "optional"},
			"AlsoSet":  FieldSpec{WireName: "also_set"},
			"Opaque":   FieldSpec{WireName: "opaque"},
		},
	}
	raw := serializeConfig(binding.Fields, nilProbeConfig{
		Required: "yes",
		AlsoSet:  Ptr(true),
	}, "probe.thing.x")
	got, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("expected a map, got %T", raw)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := got["optional"]; present {
		t.Fatalf("a nil pointer field must not appear at all, got %s", encoded)
	}
	if _, present := got["opaque"]; present {
		t.Fatalf("a nil interface field must not appear at all, got %s", encoded)
	}
	if got["also_set"] != true {
		t.Fatalf("a set pointer field must serialize to its value, got %s", encoded)
	}
	if got["required"] != "yes" {
		t.Fatalf("a required field must serialize, got %s", encoded)
	}
}

// Ptr exists because Go cannot take the address of a literal, so every
// call site would otherwise define its own one-line generic.
func TestPtr(t *testing.T) {
	if got := Ptr(60); got == nil || *got != 60 {
		t.Fatalf("Ptr(60) = %v", got)
	}
	if got := Ptr("s"); got == nil || *got != "s" {
		t.Fatalf("Ptr(\"s\") = %v", got)
	}
}
