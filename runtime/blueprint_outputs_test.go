package sdk

import (
	"encoding/json"
	"testing"
)

// blueprint_outputs_test.go covers UBI-261's runtime half: a blueprint
// caller reports what the blueprint returned, because the address an
// output refers to is knowable only while the blueprint runs.

func TestBlueprintOutputs_RecordedOnTheDocument(t *testing.T) {
	def := Stack("payments", func() {
		Intent(IntentInfo{Summary: "call a blueprint"})
		q := Resource(
			ResourceBinding{WireType: "fake_queue", Fields: FieldMap{"Name": {WireName: "name"}}},
			"orders",
			struct {
				Name string `json:"name"`
			}{Name: "orders"},
		)
		BlueprintOutputs(map[string]*Computed{
			"queue_url": q.Field("url"),
			"queue_arn": q.Field("arn"),
		})
	})

	doc, err := def.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		BlueprintOutputs map[string]string `json:"blueprint_outputs"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if got := parsed.BlueprintOutputs["queue_url"]; got != "payments.fake_queue.orders.url" {
		t.Errorf("queue_url = %q", got)
	}
	if got := parsed.BlueprintOutputs["queue_arn"]; got != "payments.fake_queue.orders.arn" {
		t.Errorf("queue_arn = %q", got)
	}
}

// Every program that is not a blueprint caller has to serialize
// byte-identically to before this existed.
func TestBlueprintOutputs_OmittedWhenNeverCalled(t *testing.T) {
	def := Stack("payments", func() {
		Intent(IntentInfo{Summary: "an ordinary program"})
	})
	doc, err := def.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatal(err)
	}
	if _, present := probe["blueprint_outputs"]; present {
		t.Errorf("blueprint_outputs must be absent, not empty: %s", raw)
	}
}

// A nil entry is skipped rather than recorded as an empty address. The
// caller reports a declared output that came back with no value as its
// own named error, where it can say which output and which blueprint;
// an empty address here would instead look like a resolved one.
func TestBlueprintOutputs_NilIsSkipped(t *testing.T) {
	def := Stack("payments", func() {
		Intent(IntentInfo{Summary: "a blueprint that returned nothing for one output"})
		q := Resource(
			ResourceBinding{WireType: "fake_queue", Fields: FieldMap{"Name": {WireName: "name"}}},
			"orders",
			struct {
				Name string `json:"name"`
			}{Name: "orders"},
		)
		BlueprintOutputs(map[string]*Computed{"queue_url": q.Field("url"), "missing": nil})
	})
	doc, err := def.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	if _, present := doc.BlueprintOutputs["missing"]; present {
		t.Error("a nil output was recorded")
	}
	if len(doc.BlueprintOutputs) != 1 {
		t.Errorf("outputs = %v, want only the real one", doc.BlueprintOutputs)
	}
}

// A caller that invokes more than one blueprint merges rather than
// overwriting.
func TestBlueprintOutputs_MergesAcrossCalls(t *testing.T) {
	def := Stack("payments", func() {
		Intent(IntentInfo{Summary: "two blueprints"})
		q := Resource(
			ResourceBinding{WireType: "fake_queue", Fields: FieldMap{"Name": {WireName: "name"}}},
			"orders",
			struct {
				Name string `json:"name"`
			}{Name: "orders"},
		)
		BlueprintOutputs(map[string]*Computed{"first": q.Field("url")})
		BlueprintOutputs(map[string]*Computed{"second": q.Field("arn")})
	})
	doc, err := def.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.BlueprintOutputs) != 2 {
		t.Errorf("outputs = %v, want both calls' own entries", doc.BlueprintOutputs)
	}
}
