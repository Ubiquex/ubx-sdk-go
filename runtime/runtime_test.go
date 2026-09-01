package sdk

import (
	"encoding/json"
	"strings"
	"testing"
)

var widgetBinding = ResourceBinding{
	WireType: "fake_widget",
	Fields: FieldMap{
		"Name":  {WireName: "name"},
		"Count": {WireName: "count"},
		"Tags":  {WireName: "tags"},
		"Owner": {WireName: "owner_ref"},
		"Timeouts": {
			WireName: "timeouts",
			Kind:     "object",
			Fields: FieldMap{
				"Create": {WireName: "create"},
			},
		},
	},
}

type widgetConfig struct {
	Name     any
	Count    any
	Tags     any
	Owner    any
	Timeouts any
}

var dataWidgetBinding = DataSourceBinding{
	WireType: "data_fake_widget",
	Fields: FieldMap{
		"ID":   {WireName: "id"},
		"Name": {WireName: "name"},
	},
}

type dataWidgetLookup struct {
	ID   any
	Name any
}

type widgetTimeouts struct {
	Create any
}

func evalJSON(t *testing.T, def *StackDefinition) map[string]any {
	t.Helper()
	doc, err := def.Evaluate()
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

func TestBasicResourceRoundTrip(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "a summary"})
		Resource(widgetBinding, "widget-a", widgetConfig{
			Name:  "hello",
			Count: 3,
		})
	})

	out := evalJSON(t, def)
	if out["schema_version"] != float64(1) {
		t.Fatalf("schema_version: %v", out["schema_version"])
	}
	if out["kind"] != "ubx:intent/v1" {
		t.Fatalf("kind: %v", out["kind"])
	}
	resources := out["resources"].([]any)
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	res := resources[0].(map[string]any)
	if res["type"] != "fake_widget" || res["name"] != "widget-a" || res["op"] != "create" {
		t.Fatalf("unexpected resource shape: %v", res)
	}
	cfg := res["config"].(map[string]any)
	if cfg["name"] != "hello" || cfg["count"] != float64(3) {
		t.Fatalf("unexpected config: %v", cfg)
	}
	if _, present := cfg["owner_ref"]; present {
		t.Fatalf("unset field leaked into config: %v", cfg)
	}
}

func TestComputedReferenceSerializesToRef(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		first := Resource(widgetBinding, "a", widgetConfig{Name: "a"})
		Resource(widgetBinding, "b", widgetConfig{Owner: first.Field("id")})
	})

	out := evalJSON(t, def)
	resources := out["resources"].([]any)
	second := resources[1].(map[string]any)
	cfg := second["config"].(map[string]any)
	ownerRef := cfg["owner_ref"].(map[string]any)
	ref := ownerRef["$ref"].(map[string]any)
	if ref["to"] != "demo.fake_widget.a.id" {
		t.Fatalf("unexpected $ref: %v", ref)
	}
}

func TestSecretAndCrossMarkers(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{
			Name:  Secret("vault", "path/to/secret"),
			Owner: Cross("../networking", "networking.fake_widget.shared"),
		})
	})

	out := evalJSON(t, def)
	cfg := out["resources"].([]any)[0].(map[string]any)["config"].(map[string]any)
	secret := cfg["name"].(map[string]any)["$secret"].(map[string]any)
	if secret["backend"] != "vault" || secret["path"] != "path/to/secret" {
		t.Fatalf("unexpected $secret: %v", secret)
	}
	cross := cfg["owner_ref"].(map[string]any)["$cross"].(map[string]any)
	if cross["ledger_dir"] != "../networking" || cross["to"] != "networking.fake_widget.shared" {
		t.Fatalf("unexpected $cross: %v", cross)
	}
}

func TestNestedObjectField(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{
			Name:     "a",
			Timeouts: widgetTimeouts{Create: "20m"},
		})
	})

	out := evalJSON(t, def)
	cfg := out["resources"].([]any)[0].(map[string]any)["config"].(map[string]any)
	timeouts := cfg["timeouts"].(map[string]any)
	if timeouts["create"] != "20m" {
		t.Fatalf("unexpected nested object: %v", timeouts)
	}
}

func TestOpaqueMapField(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{
			Name: "a",
			Tags: map[string]any{"env": "prod", "owner": "roozbeh"},
		})
	})

	out := evalJSON(t, def)
	cfg := out["resources"].([]any)[0].(map[string]any)["config"].(map[string]any)
	tags := cfg["tags"].(map[string]any)
	if tags["env"] != "prod" || tags["owner"] != "roozbeh" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestMissingIntentIsHardFailure(t *testing.T) {
	def := Stack("demo", func() {
		Resource(widgetBinding, "a", widgetConfig{Name: "a"})
	})
	_, err := def.Evaluate()
	if err == nil || !strings.Contains(err.Error(), "Intent() was never called") {
		t.Fatalf("expected missing-intent error, got %v", err)
	}
}

func TestDuplicateResourceAddressIsHardFailure(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{Name: "1"})
		Resource(widgetBinding, "a", widgetConfig{Name: "2"})
	})
	_, err := def.Evaluate()
	if err == nil || !strings.Contains(err.Error(), "duplicate resource") {
		t.Fatalf("expected duplicate-resource error, got %v", err)
	}
}

func TestUnrecognizedConfigFieldIsHardFailure(t *testing.T) {
	type badConfig struct {
		NotARealField any
	}
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", badConfig{NotARealField: "x"})
	})
	_, err := def.Evaluate()
	if err == nil || !strings.Contains(err.Error(), "unrecognized config field") {
		t.Fatalf("expected unrecognized-field error, got %v", err)
	}
}

func TestFunctionInConfigIsHardFailure(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{Name: func() {}})
	})
	_, err := def.Evaluate()
	if err == nil || !strings.Contains(err.Error(), "cannot appear in a resource's own config") {
		t.Fatalf("expected function-rejection error, got %v", err)
	}
}

func TestResourceOutsideStackIsHardFailure(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "called outside of an active Stack") {
			t.Fatalf("expected panic about calling outside Stack(), got %v", r)
		}
	}()
	Resource(widgetBinding, "a", widgetConfig{Name: "a"})
}

// --- Data() -- mirrors the Resource() tests above exactly (same
// duplicate-address check, same marker-aware serializer, same
// blueprint-provenance wiring) rather than a separate, narrower suite,
// since addDataSource's own doc comment states it reuses those same
// mechanisms unchanged.

func TestBasicDataSourceRoundTrip(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "a summary"})
		Data(dataWidgetBinding, "widget-a", dataWidgetLookup{
			ID: "w-123",
		})
	})

	out := evalJSON(t, def)
	dataSources := out["data_sources"].([]any)
	if len(dataSources) != 1 {
		t.Fatalf("expected 1 data source, got %d", len(dataSources))
	}
	ds := dataSources[0].(map[string]any)
	if ds["type"] != "data_fake_widget" || ds["name"] != "widget-a" {
		t.Fatalf("unexpected data source shape: %v", ds)
	}
	if _, present := ds["op"]; present {
		t.Fatalf("data_sources[] entry must not carry an op field, got: %v", ds)
	}
	lookup := ds["lookup"].(map[string]any)
	if lookup["id"] != "w-123" {
		t.Fatalf("unexpected lookup: %v", lookup)
	}
	if _, present := lookup["name"]; present {
		t.Fatalf("unset field leaked into lookup: %v", lookup)
	}
}

// TestDataSourceResultFeedsResourceConfig is the design's own core
// requirement: Data() returns the identical *Computed handle Resource()
// does, so its result wires into a sibling resource's config through
// the exact same $ref serialization -- no separate reference type.
func TestDataSourceResultFeedsResourceConfig(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		looked := Data(dataWidgetBinding, "existing", dataWidgetLookup{ID: "w-123"})
		Resource(widgetBinding, "b", widgetConfig{Owner: looked.Field("name")})
	})

	out := evalJSON(t, def)
	resources := out["resources"].([]any)
	cfg := resources[0].(map[string]any)["config"].(map[string]any)
	ownerRef := cfg["owner_ref"].(map[string]any)
	ref := ownerRef["$ref"].(map[string]any)
	if ref["to"] != "demo.data_fake_widget.existing.name" {
		t.Fatalf("unexpected $ref into a data source's own result: %v", ref)
	}
}

// TestResourceComputedFeedsDataSourceLookup is the reverse direction of
// the same requirement: a data source's own lookup can reference a
// sibling resource's computed output (e.g. looking an instance up by an
// ID a same-batch create just produced) -- the identical $ref
// resolution a resource's config already gets, reused unchanged.
func TestResourceComputedFeedsDataSourceLookup(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		created := Resource(widgetBinding, "a", widgetConfig{Name: "a"})
		Data(dataWidgetBinding, "lookup-created", dataWidgetLookup{ID: created.Field("id")})
	})

	out := evalJSON(t, def)
	dataSources := out["data_sources"].([]any)
	lookup := dataSources[0].(map[string]any)["lookup"].(map[string]any)
	idRef := lookup["id"].(map[string]any)["$ref"].(map[string]any)
	if idRef["to"] != "demo.fake_widget.a.id" {
		t.Fatalf("unexpected $ref in a data source's own lookup: %v", idRef)
	}
}

func TestDuplicateDataSourceAddressIsHardFailure(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Data(dataWidgetBinding, "a", dataWidgetLookup{ID: "1"})
		Data(dataWidgetBinding, "a", dataWidgetLookup{ID: "2"})
	})
	_, err := def.Evaluate()
	if err == nil || !strings.Contains(err.Error(), "duplicate data source") {
		t.Fatalf("expected duplicate-data-source error, got %v", err)
	}
}

func TestDataSourceMissingNameIsHardFailure(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Data(dataWidgetBinding, "", dataWidgetLookup{ID: "1"})
	})
	_, err := def.Evaluate()
	if err == nil || !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("expected missing-name error, got %v", err)
	}
}

func TestDataSourceOutsideStackIsHardFailure(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "called outside of an active Stack") {
			t.Fatalf("expected panic about calling outside Stack(), got %v", r)
		}
	}()
	Data(dataWidgetBinding, "a", dataWidgetLookup{ID: "1"})
}

// TestNoDataSourcesOmitsFieldEntirely: a stack that never calls Data()
// must not carry a "data_sources" key at all -- the overwhelming
// majority of existing programs, zero wire-format change for them.
func TestNoDataSourcesOmitsFieldEntirely(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{Name: "a"})
	})

	out := evalJSON(t, def)
	if _, present := out["data_sources"]; present {
		t.Fatalf("expected no data_sources key when Data() was never called, got: %v", out["data_sources"])
	}
}

// blueprintWidgetBinding (UBI-225) is widgetBinding's own real sibling,
// differing in exactly one field: BlueprintName is set, matching what
// blueprint/gogen.go's own renderGoBindings now stamps into a real
// blueprint's own generated bindings.go. Used by every test below that
// needs a binding a blueprint actually produced, as opposed to an
// ordinary provider SDK binding (sdk/codegen/templates/go never sets
// this field, so widgetBinding above stays a correct stand-in for that
// case).
var blueprintWidgetBinding = ResourceBinding{
	WireType:      "fake_widget",
	Fields:        FieldMap{"Name": {WireName: "name"}},
	BlueprintName: "ci-platform",
}

func resourceSources(t *testing.T, res map[string]any) []any {
	t.Helper()
	sources, _ := res["sources"].([]any)
	return sources
}

// TestBlueprintBinding_StampsProvenanceWithoutOpenScope is UBI-225's own
// required proof: a resource built by calling Resource() directly
// against a blueprint's own exported binding -- never through that
// blueprint's own generated wrapper function, so PushBlueprintSource is
// never called at all -- must still carry a real blueprint source. Before
// this fix, importing a blueprint's own bindings.go directly and
// constructing a resource by hand produced zero provenance, silently
// indistinguishable from an ordinary hand-written resource.
func TestBlueprintBinding_StampsProvenanceWithoutOpenScope(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(blueprintWidgetBinding, "bypass", widgetConfig{Name: "bypass-widget"})
	})

	out := evalJSON(t, def)
	res := out["resources"].([]any)[0].(map[string]any)
	sources := resourceSources(t, res)
	if len(sources) != 1 {
		t.Fatalf("expected exactly 1 source, got %d: %v", len(sources), res)
	}
	src := sources[0].(map[string]any)
	if src["kind"] != "blueprint" || src["ref"] != "ci-platform" {
		t.Fatalf("expected {kind: blueprint, ref: ci-platform}, got %v", src)
	}
}

// TestOrdinaryBinding_NoBlueprintName_NoProvenance confirms zero
// regression: an ordinary provider SDK binding (BlueprintName never
// set) still produces no "sources" key at all, exactly as before this
// fix -- the overwhelming majority of real resources.
func TestOrdinaryBinding_NoBlueprintName_NoProvenance(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		Resource(widgetBinding, "a", widgetConfig{Name: "a"})
	})

	out := evalJSON(t, def)
	res := out["resources"].([]any)[0].(map[string]any)
	if _, present := res["sources"]; present {
		t.Fatalf("expected no sources key for an ordinary binding, got: %v", res["sources"])
	}
}

// TestPushBlueprintSource_WinsOverBindingBlueprintName confirms the real
// priority order when both signals are present at once: an open
// PushBlueprintSource scope always wins over whatever BlueprintName the
// binding itself carries. Every real call blueprint/gogen.go's own
// generated code makes has both signals agree (a wrapper function's own
// Resource() calls run against that SAME blueprint's own bindings), so
// this case is never actually reachable through generated code today --
// pinned here anyway, as real, deliberate behavior, not left unspecified.
func TestPushBlueprintSource_WinsOverBindingBlueprintName(t *testing.T) {
	def := Stack("demo", func() {
		Intent(IntentInfo{Summary: "s"})
		PushBlueprintSource("outer-blueprint")
		Resource(blueprintWidgetBinding, "nested", widgetConfig{Name: "nested-widget"})
		PopBlueprintSource()
	})

	out := evalJSON(t, def)
	res := out["resources"].([]any)[0].(map[string]any)
	sources := resourceSources(t, res)
	if len(sources) != 1 {
		t.Fatalf("expected exactly 1 source, got %d: %v", len(sources), res)
	}
	src := sources[0].(map[string]any)
	if src["ref"] != "outer-blueprint" {
		t.Fatalf("expected the open scope (%q) to win over the binding's own BlueprintName (%q), got ref %v", "outer-blueprint", blueprintWidgetBinding.BlueprintName, src["ref"])
	}
}
