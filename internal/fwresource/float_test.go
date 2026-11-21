package fwresource

import (
	"context"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A number from NetBox must equal the same number written in the configuration, otherwise every
// plan after an import shows an update that has no visible change.
func TestFloatFromAPIEqualsFloatFromConfiguration(t *testing.T) {
	for _, text := range []string{"40.37", "-2.49", "0.25", "40", "-3", "1.5", "0.1", "123456.789"} {
		// this is how Terraform hands a configured number to the provider
		bf, _, err := big.ParseFloat(text, 10, 512, big.ToNearestEven)
		if err != nil {
			t.Fatal(err)
		}
		fromConfig, err := types.Float64Type.ValueFromTerraform(context.Background(), tftypes.NewValue(tftypes.Number, bf))
		if err != nil {
			t.Fatal(err)
		}
		n, _ := strconv.ParseFloat(text, 64)
		if got := float64Value(n); !got.Equal(fromConfig) {
			t.Errorf("%s from the API does not equal %s from the configuration", got, fromConfig)
		}
	}
}

// An empty document is the same as none, except for an attribute whose default is an empty document:
// its state must hold what the plan holds, or an import is followed by an update.
func TestEmptyJSONDocumentAndDefault(t *testing.T) {
	plain := Attr{Name: "p", Kind: JSON, Nullable: true, Optional: true}
	if got := plain.fromJSON(map[string]any{}, nil); !got.IsNull() {
		t.Errorf("an empty document without a default read as %v, want null", got)
	}
	withDefault := Attr{Name: "d", Kind: JSON, Nullable: true, Optional: true, Default: "{}"}
	if got := withDefault.fromJSON(map[string]any{}, nil); got.IsNull() || got.String() != `"{}"` {
		t.Errorf("an empty document with the default {} read as %v, want {}", got)
	}
}

// A value an attribute has while it is not set must not count as asking for a newer NetBox: the state
// holds it and so does a configuration written from the state.
func TestIsConfiguredIgnoresUnsetValues(t *testing.T) {
	zero := Attr{Name: "owner_id", Kind: Int, Optional: true, zero: true}
	if isConfigured(zero, types.Int64Value(0)) {
		t.Error("0 of an attribute that reads as 0 while not set counted as configured")
	}
	if !isConfigured(zero, types.Int64Value(5)) {
		t.Error("a real value did not count")
	}
	withDefault := Attr{Name: "auto_sync_enabled", Kind: Bool, Optional: true, Default: false}
	if isConfigured(withDefault, types.BoolValue(false)) || !isConfigured(withDefault, types.BoolValue(true)) {
		t.Error("the default of an attribute")
	}
	if isConfigured(zero, types.Int64Null()) {
		t.Error("null counted as configured")
	}
}

// A text written as a JSON string, with its quotes, is the same text when NetBox returns it without them.
func TestLooseJSONKeepsAQuotedString(t *testing.T) {
	a := Attr{Name: "default", Kind: String, LooseJSON: true}
	prior := types.StringValue(`"abc"`)
	if got := a.fromJSON("abc", prior); !got.Equal(prior) {
		t.Errorf("got %v, want the prior %v", got, prior)
	}
	if got := a.fromJSON("other", prior); got.Equal(prior) {
		t.Error("a different text kept the prior value")
	}
}

func TestDefaultIsZero(t *testing.T) {
	for d, want := range map[any]bool{false: true, true: false, "": true, "x": false, int64(0): true, int64(2): false, 0: true, 3: false, 0.0: true, nil: false} {
		if got := defaultIsZero(d); got != want {
			t.Errorf("defaultIsZero(%v) = %v, want %v", d, got, want)
		}
	}
}

func TestFloat64ValueKeepsInfinity(t *testing.T) {
	if v := float64Value(math.Inf(1)); !math.IsInf(v.ValueFloat64(), 1) {
		t.Errorf("infinity became %v", v)
	}
}

// A value that equals the default of its attribute is not something the configuration asks for.
func TestIsConfiguredIgnoresTheDefaultOfEveryKind(t *testing.T) {
	for name, tc := range map[string]struct {
		a       Attr
		same    attr.Value
		differs attr.Value
	}{
		"bool":   {Attr{Kind: Bool, Optional: true, Default: true}, types.BoolValue(true), types.BoolValue(false)},
		"int":    {Attr{Kind: Int, Optional: true, Default: int64(5)}, types.Int64Value(5), types.Int64Value(6)},
		"float":  {Attr{Kind: Float, Optional: true, Default: 1.5}, types.Float64Value(1.5), types.Float64Value(2.5)},
		"string": {Attr{Kind: String, Optional: true, Default: "x"}, types.StringValue("x"), types.StringValue("y")},
	} {
		if isConfigured(tc.a, tc.same) {
			t.Errorf("%s: the default counted as configured", name)
		}
		if !isConfigured(tc.a, tc.differs) {
			t.Errorf("%s: another value did not count", name)
		}
	}
}
