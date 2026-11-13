package fwresource

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringFromJSON turns a value of the API into a string. Choices come as objects with a value and a label.
func stringFromJSON(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", false
	case string:
		return x, true
	case bool:
		return strconv.FormatBool(x), true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case map[string]any:
		if val, ok := x["value"]; ok {
			return stringFromJSON(val)
		}
		if id, ok := x["id"]; ok {
			return stringFromJSON(id)
		}
		if b, err := json.Marshal(x); err == nil {
			return string(b), true
		}
	default:
		if b, err := json.Marshal(x); err == nil {
			return string(b), true
		}
	}
	return "", false
}

// intFromJSON turns a value of the API into a number. Nested objects count as their ID, choices as
// their value.
func intFromJSON(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(math.Round(x)), true
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return n, true
		}
	case map[string]any:
		if id, ok := x["id"]; ok {
			return intFromJSON(id)
		}
		if val, ok := x["value"]; ok {
			return intFromJSON(val)
		}
	}
	return 0, false
}

func floatFromJSON(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		if n, err := strconv.ParseFloat(x, 64); err == nil {
			return n, true
		}
	}
	return 0, false
}

func boolFromJSON(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case map[string]any:
		if val, ok := x["value"]; ok {
			return boolFromJSON(val)
		}
	}
	return false, false
}

// fromJSON converts a field of the API response into the value of an attribute. prior is the value
// the attribute had in the plan or state, and tells how to represent an empty field: an optional
// attribute that was not set stays unset instead of becoming an empty string.
func (a Attr) fromJSON(v any, prior attr.Value) attr.Value {
	priorSet := prior != nil && !prior.IsNull() && !prior.IsUnknown()
	keepNull := !priorSet && !a.Computed && !a.Required && a.Default == nil

	switch a.Kind {
	case String:
		if a.IPOnly {
			addr, _ := v.(string)
			if ip, _, err := net.ParseCIDR(addr); err == nil {
				return types.StringValue(ip.String())
			}
			return types.StringValue("")
		}
		if a.LooseJSON {
			return a.looseJSONFromJSON(v, prior)
		}
		s, ok := stringFromJSON(v)
		if ok && priorSet {
			if ps, isStr := prior.(types.String); isStr && strings.TrimSpace(ps.ValueString()) == strings.TrimSpace(s) {
				return prior
			}
			if ps, isStr := prior.(types.String); isStr && a.Time {
				t1, err1 := time.Parse(time.RFC3339, ps.ValueString())
				t2, err2 := time.Parse(time.RFC3339, s)
				if err1 == nil && err2 == nil && t1.Equal(t2) {
					return prior
				}
			}
		}
		if ok && a.CaseInsensitive && priorSet {
			if ps, isStr := prior.(types.String); isStr && strings.EqualFold(ps.ValueString(), s) {
				return prior
			}
		}
		if !ok || s == "" {
			if ok && priorSet {
				return types.StringValue("")
			}
			if keepNull || (!ok && !a.Computed) {
				return types.StringNull()
			}
			return types.StringValue("")
		}
		if a.Time {
			// A timestamp is written the same way whatever NetBox returns, with milliseconds.
			if parsed, err := time.Parse(time.RFC3339, s); err == nil {
				s = parsed.UTC().Format("2006-01-02T15:04:05.000Z")
			}
		}
		return types.StringValue(s)
	case JSON:
		if v == nil {
			return types.StringNull()
		}
		if !priorSet {
			// an empty document is the same as none
			if m, ok := v.(map[string]any); ok && len(m) == 0 {
				return types.StringNull()
			}
			if l, ok := v.([]any); ok && len(l) == 0 {
				return types.StringNull()
			}
		}
		b, _ := json.Marshal(v)
		if priorSet {
			if ps, ok := prior.(types.String); ok && jsonEqual(ps.ValueString(), string(b)) {
				return prior
			}
		}
		return types.StringValue(string(b))
	case Int:
		n, ok := intFromJSON(v)
		if !ok {
			if a.Default != nil {
				return types.Int64Value(toInt64(a.Default))
			}
			if a.zero || a.Zero {
				return types.Int64Value(0)
			}
			return types.Int64Null()
		}
		return types.Int64Value(n)
	case Float:
		n, ok := floatFromJSON(v)
		if !ok {
			if a.zero || a.Zero {
				return types.Float64Value(0)
			}
			return types.Float64Null()
		}
		return types.Float64Value(n)
	case Bool:
		b, ok := boolFromJSON(v)
		if !ok {
			if d, isBool := a.Default.(bool); isBool {
				return types.BoolValue(d)
			}
			if a.zero || a.Zero {
				return types.BoolValue(false)
			}
			return types.BoolNull()
		}
		if !b && keepNull {
			return types.BoolNull()
		}
		return types.BoolValue(b)
	case StringMap:
		m, _ := v.(map[string]any)
		elems := map[string]attr.Value{}
		for k, e := range m {
			if str := customFieldValueToString(e); str != "" {
				elems[k] = types.StringValue(str)
			}
		}
		out, _ := types.MapValue(types.StringType, elems)
		return out
	case NestedSet, NestedList:
		list, _ := v.([]any)
		if m, ok := v.(map[string]any); ok {
			// a single nested object is a list of one
			list = []any{m}
		}
		if a.Pairs {
			conv := make([]any, 0, len(list))
			for _, e := range list {
				if pair, ok := e.([]any); ok && len(pair) >= 2 {
					conv = append(conv, map[string]any{"start": pair[0], "end": pair[1]})
				}
			}
			list = conv
		}
		if len(list) == 0 && !priorSet && !a.Computed {
			return a.nullCollection()
		}
		elems := make([]attr.Value, 0, len(list))
		ot := a.objectType()
		for _, e := range list {
			m, _ := e.(map[string]any)
			vals := map[string]attr.Value{}
			for _, sub := range a.Sub {
				sub.Computed = true
				sub.Zero = true
				sv := sub.fromJSON(sub.pick(m), nil)
				if sv == nil || sv.IsNull() && sub.Required {
					sv = nullOf(sub)
				}
				vals[sub.Name] = sv
			}
			obj, diags := types.ObjectValue(ot.AttrTypes, vals)
			if !diags.HasError() {
				elems = append(elems, obj)
			}
		}
		return a.collection(elems)
	case IntLists, StringLists:
		list, _ := v.([]any)
		if len(list) == 0 && keepNull {
			return a.nullCollection()
		}
		inner := a.elemType().(types.ListType).ElemType
		elems := make([]attr.Value, 0, len(list))
		for _, e := range list {
			items, _ := e.([]any)
			vals := make([]attr.Value, 0, len(items))
			for _, it := range items {
				if a.Kind == IntLists {
					if n, ok := intFromJSON(it); ok {
						vals = append(vals, types.Int64Value(n))
					}
				} else if s, ok := stringFromJSON(it); ok {
					vals = append(vals, types.StringValue(s))
				}
			}
			l, _ := types.ListValue(inner, vals)
			elems = append(elems, l)
		}
		return a.collection(elems)
	case StringSet, StringList:
		list, _ := v.([]any)
		if len(list) == 0 && keepNull {
			return a.nullCollection()
		}
		elems := make([]attr.Value, 0, len(list))
		for _, e := range list {
			e = a.elem(e)
			if s, ok := stringFromJSON(e); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		return a.collection(elems)
	case IntSet, IntList:
		list, _ := v.([]any)
		if len(list) == 0 && keepNull {
			return a.nullCollection()
		}
		elems := make([]attr.Value, 0, len(list))
		for _, e := range list {
			e = a.elem(e)
			if n, ok := intFromJSON(e); ok {
				elems = append(elems, types.Int64Value(n))
			}
		}
		return a.collection(elems)
	}
	return nil
}

func (a Attr) elemType() attr.Type {
	switch a.Kind {
	case IntSet, IntList:
		return types.Int64Type
	case IntLists:
		return types.ListType{ElemType: types.Int64Type}
	case StringLists:
		return types.ListType{ElemType: types.StringType}
	case NestedSet, NestedList:
		return a.objectType()
	}
	return types.StringType
}

func (a Attr) isSet() bool {
	return a.Kind == StringSet || a.Kind == IntSet || a.Kind == NestedSet
}

// objectType is the type of the objects of a nested attribute.
func (a Attr) objectType() types.ObjectType {
	t := map[string]attr.Type{}
	for _, sub := range a.Sub {
		t[sub.Name] = sub.valueType()
	}
	return types.ObjectType{AttrTypes: t}
}

func (a Attr) nullCollection() attr.Value {
	if a.isSet() {
		return types.SetNull(a.elemType())
	}
	return types.ListNull(a.elemType())
}

func (a Attr) collection(elems []attr.Value) attr.Value {
	if a.isSet() {
		s, _ := types.SetValue(a.elemType(), elems)
		return s
	}
	l, _ := types.ListValue(a.elemType(), elems)
	return l
}

// toJSON converts the value of an attribute into the field to send. The second result is false
// when the field is left out. A null value clears the field when clear is set (on update).
func (a Attr) toJSON(val attr.Value, prior attr.Value) (any, bool, error) {
	clear := prior != nil
	if val == nil || val.IsUnknown() {
		return nil, false, nil
	}
	if a.zero && isZero(val) && a.Nullable && a.Kind == Int && !a.OmitZero {
		// NetBox wants an explicit null for a field that is not set, and it clears the field on update.
		return nil, true, nil
	}
	if val.IsNull() && a.Computed && !a.zero {
		// a value NetBox computes is never cleared
		return nil, false, nil
	}
	if a.zero && isZero(val) {
		// An attribute that is not set. It is left out on create, and cleared on update when it was set.
		switch {
		case (a.Kind == Int && !a.Ref && !a.OmitZero), a.Kind == Bool:
			// numbers and booleans are sent as they are
		case !clear || (prior != nil && isZero(prior)):
			return nil, false, nil
		default:
			val = types.StringNull()
			if a.Kind != String {
				val = types.Int64Null()
			}
		}
	}
	if a.Ref && a.Kind == Int {
		if n, ok := val.(types.Int64); ok && !n.IsNull() && n.ValueInt64() == 0 {
			val = types.Int64Null()
			if !clear {
				return nil, false, nil
			}
		}
	}
	if val.IsNull() {
		if !clear {
			return nil, false, nil
		}
		switch a.Kind {
		case String:
			if a.Nullable {
				return nil, true, nil
			}
			return "", true, nil
		case StringSet, StringList, IntSet, IntList, IntLists, StringLists:
			return []any{}, true, nil
		}
		return nil, true, nil
	}

	switch v := val.(type) {
	case types.String:
		if a.LooseJSON {
			var out any
			if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
				return v.ValueString(), true, nil
			}
			return out, true, nil
		}
		if a.Kind == JSON {
			var out any
			if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
				return nil, false, fmt.Errorf("attribute %s is not valid JSON: %w", a.Name, err)
			}
			return out, true, nil
		}
		return v.ValueString(), true, nil
	case types.Int64:
		return v.ValueInt64(), true, nil
	case types.Float64:
		return v.ValueFloat64(), true, nil
	case types.Bool:
		return v.ValueBool(), true, nil
	case types.Set:
		return a.collectionJSON(v.Elements()), true, nil
	case types.List:
		return a.collectionJSON(v.Elements()), true, nil
	}
	return nil, false, fmt.Errorf("attribute %s has an unsupported value", a.Name)
}

func (a Attr) collectionJSON(elems []attr.Value) []any {
	out := make([]any, 0, len(elems))
	for _, e := range elems {
		switch v := e.(type) {
		case types.Object:
			m := map[string]any{}
			for _, sub := range a.Sub {
				val, ok, err := sub.toJSON(v.Attributes()[sub.Name], nil)
				if err == nil && ok {
					m[sub.key()] = val
				}
			}
			out = append(out, m)
		case types.String:
			out = append(out, v.ValueString())
		case types.Int64:
			out = append(out, v.ValueInt64())
		case types.List:
			inner := make([]any, 0, len(v.Elements()))
			for _, it := range v.Elements() {
				switch x := it.(type) {
				case types.String:
					inner = append(inner, x.ValueString())
				case types.Int64:
					inner = append(inner, x.ValueInt64())
				}
			}
			out = append(out, inner)
		}
	}
	return out
}

func jsonEqual(a, b string) bool {
	var ja, jb any
	if json.Unmarshal([]byte(a), &ja) != nil || json.Unmarshal([]byte(b), &jb) != nil {
		return strings.TrimSpace(a) == strings.TrimSpace(b)
	}
	x, _ := json.Marshal(ja)
	y, _ := json.Marshal(jb)
	return string(x) == string(y)
}

// isZero tells whether a value is the zero value of its type: an empty string, 0, false or an empty
// collection. Null counts too.
func isZero(v attr.Value) bool {
	if v == nil || v.IsNull() {
		return true
	}
	switch x := v.(type) {
	case types.String:
		return x.ValueString() == ""
	case types.Int64:
		return x.ValueInt64() == 0
	case types.Float64:
		return x.ValueFloat64() == 0
	case types.Bool:
		return !x.ValueBool()
	case types.Set:
		return len(x.Elements()) == 0
	case types.List:
		return len(x.Elements()) == 0
	}
	return false
}

// looseJSONFromJSON turns a field that is a string or any JSON value into a string. The string of
// the configuration is kept when it means the same.
func (a Attr) looseJSONFromJSON(v any, prior attr.Value) attr.Value {
	if v == nil {
		return types.StringValue("")
	}
	var s string
	if str, ok := v.(string); ok {
		s = str
	} else {
		b, _ := json.Marshal(v)
		s = string(b)
	}
	if ps, ok := prior.(types.String); ok && !ps.IsNull() && !ps.IsUnknown() {
		if ps.ValueString() == s || jsonEqual(ps.ValueString(), s) {
			return prior
		}
	}
	return types.StringValue(s)
}

// elem picks the field named by ElemKey from an element that is an object.
func (a Attr) elem(e any) any {
	if a.ElemKey == "" {
		return e
	}
	if m, ok := e.(map[string]any); ok {
		return m[a.ElemKey]
	}
	return e
}

// valueType is the type of the values of the attribute.
func (a Attr) valueType() attr.Type {
	switch a.Kind {
	case Int:
		return types.Int64Type
	case Bool:
		return types.BoolType
	case Float:
		return types.Float64Type
	case StringMap:
		return types.MapType{ElemType: types.StringType}
	case StringSet, IntSet:
		return types.SetType{ElemType: a.elemType()}
	case StringList, IntList, IntLists, StringLists:
		return types.ListType{ElemType: a.elemType()}
	case NestedSet:
		return types.SetType{ElemType: a.objectType()}
	case NestedList:
		return types.ListType{ElemType: a.objectType()}
	}
	return types.StringType
}

// pick returns the field of the object the attribute is read from.
func (a Attr) pick(obj map[string]any) any {
	if len(a.Path) == 0 {
		return obj[a.key()]
	}
	var cur any = obj
	for _, seg := range a.Path {
		switch x := cur.(type) {
		case map[string]any:
			cur = x[seg]
		case []any:
			if seg == "[0]" && len(x) > 0 {
				cur = x[0]
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return cur
}
