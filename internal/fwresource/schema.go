package fwresource

import (
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	tagsKey         = "tags"
	tagsAllKey      = "tags_all"
	customFieldsKey = "custom_fields"
)

const tagsDescription = "Tags assigned to the object, by name."
const tagsAllDescription = "All tags assigned to the object, including those not managed by Terraform."
const customFieldsDescription = "Custom field values of the object, as a map of field name to value."

// buildSchema creates the framework schema of a spec.
func (s *Spec) buildSchema() schema.Schema {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}

	blocks := map[string]schema.Block{}
	for _, a := range s.Attrs {
		switch {
		case (a.Kind == NestedSet || a.Kind == NestedList) && !a.Computed:
			blocks[a.Name] = a.schemaBlock()
		default:
			attrs[a.Name] = a.schemaAttribute()
		}
	}

	if s.Tags {
		attrs[tagsKey] = schema.SetAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: tagsDescription,
		}
		attrs[tagsAllKey] = schema.SetAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			MarkdownDescription: tagsAllDescription,
		}
	}
	if s.CustomFields {
		attrs[customFieldsKey] = schema.MapAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			MarkdownDescription: customFieldsDescription,
			PlanModifiers: []planmodifier.Map{
				mapplanmodifier.UseStateForUnknown(),
			},
		}
	}

	return schema.Schema{
		Version:             s.Version,
		MarkdownDescription: s.Description,
		Attributes:          attrs,
		Blocks:              blocks,
	}
}

// nestedAttributes creates the attributes of the objects in a nested attribute or block.
func (a Attr) nestedAttributes() map[string]schema.Attribute {
	out := map[string]schema.Attribute{}
	for _, sub := range a.Sub {
		sub.ForceNew = false
		out[sub.Name] = sub.schemaAttribute()
	}
	return out
}

func (a Attr) schemaBlock() schema.Block {
	var v []validator.Set
	var lv []validator.List
	if a.Required {
		v = append(v, setvalidator.SizeAtLeast(1))
		lv = append(lv, listvalidator.SizeAtLeast(1))
	}
	if a.Kind == NestedSet {
		return schema.SetNestedBlock{
			MarkdownDescription: a.Description,
			NestedObject:        schema.NestedBlockObject{Attributes: a.nestedAttributes()},
			Validators:          v,
		}
	}
	return schema.ListNestedBlock{
		MarkdownDescription: a.Description,
		NestedObject:        schema.NestedBlockObject{Attributes: a.nestedAttributes()},
		Validators:          lv,
	}
}

func (a Attr) schemaAttribute() schema.Attribute {
	computed := a.Computed || a.Default != nil
	// A computed value that the user can also set keeps its value while nothing else is asked for.
	stable := computed && (a.Optional || a.Default != nil) && !a.Volatile

	switch a.Kind {
	case String, JSON:
		out := schema.StringAttribute{
			Required:            a.Required,
			Optional:            a.Optional,
			Computed:            computed,
			Sensitive:           a.Sensitive,
			MarkdownDescription: a.Description,
			DeprecationMessage:  a.Deprecated,
		}
		if a.Default != nil {
			out.Default = stringdefault.StaticString(a.Default.(string))
		}
		if stable {
			out.PlanModifiers = append(out.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			out.PlanModifiers = append(out.PlanModifiers, stringplanmodifier.RequiresReplace())
		}
		var v []validator.String
		if len(a.OneOf) > 0 {
			v = append(v, stringvalidator.OneOf(a.OneOf...))
		}
		if a.MinLen > 0 || a.MaxLen > 0 {
			max := a.MaxLen
			if max == 0 {
				max = 1 << 30
			}
			v = append(v, stringvalidator.LengthBetween(a.MinLen, max))
		}
		if a.Regex != "" {
			v = append(v, stringvalidator.RegexMatches(regexp.MustCompile(a.Regex), "must match "+a.Regex))
		}
		out.Validators = v
		return out
	case Int:
		out := schema.Int64Attribute{
			Required:            a.Required,
			Optional:            a.Optional,
			Computed:            computed,
			Sensitive:           a.Sensitive,
			MarkdownDescription: a.Description,
			DeprecationMessage:  a.Deprecated,
		}
		if a.Default != nil {
			out.Default = int64default.StaticInt64(toInt64(a.Default))
		}
		if stable {
			out.PlanModifiers = append(out.PlanModifiers, int64planmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			out.PlanModifiers = append(out.PlanModifiers, int64planmodifier.RequiresReplace())
		}
		switch {
		case a.Min != nil && a.Max != nil:
			out.Validators = []validator.Int64{int64validator.Between(*a.Min, *a.Max)}
		case a.Min != nil:
			out.Validators = []validator.Int64{int64validator.AtLeast(*a.Min)}
		case a.Max != nil:
			out.Validators = []validator.Int64{int64validator.AtMost(*a.Max)}
		}
		return out
	case Float:
		out := schema.Float64Attribute{
			Required:            a.Required,
			Optional:            a.Optional,
			Computed:            computed,
			Sensitive:           a.Sensitive,
			MarkdownDescription: a.Description,
			DeprecationMessage:  a.Deprecated,
		}
		if a.Default != nil {
			out.Default = float64default.StaticFloat64(toFloat64(a.Default))
		}
		if stable {
			out.PlanModifiers = append(out.PlanModifiers, float64planmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			out.PlanModifiers = append(out.PlanModifiers, float64planmodifier.RequiresReplace())
		}
		if a.Min != nil || a.Max != nil {
			lo, hi := -1e18, 1e18
			if a.Min != nil {
				lo = float64(*a.Min)
			}
			if a.Max != nil {
				hi = float64(*a.Max)
			}
			out.Validators = []validator.Float64{float64validator.Between(lo, hi)}
		}
		return out
	case Bool:
		out := schema.BoolAttribute{
			Required:            a.Required,
			Optional:            a.Optional,
			Computed:            computed,
			Sensitive:           a.Sensitive,
			MarkdownDescription: a.Description,
			DeprecationMessage:  a.Deprecated,
		}
		if a.Default != nil {
			out.Default = booldefault.StaticBool(a.Default.(bool))
		}
		if stable {
			out.PlanModifiers = append(out.PlanModifiers, boolplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			out.PlanModifiers = append(out.PlanModifiers, boolplanmodifier.RequiresReplace())
		}
		return out
	case StringSet, IntSet:
		out := schema.SetAttribute{
			ElementType:         a.elemType(),
			Required:            a.Required,
			Optional:            a.Optional,
			Computed:            computed,
			Sensitive:           a.Sensitive,
			MarkdownDescription: a.Description,
			DeprecationMessage:  a.Deprecated,
		}
		if a.Default != nil {
			out.Default = setdefault.StaticValue(types.SetValueMust(a.elemType(), []attr.Value{}))
		}
		if stable {
			out.PlanModifiers = append(out.PlanModifiers, setplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			out.PlanModifiers = append(out.PlanModifiers, setplanmodifier.RequiresReplace())
		}
		if len(a.ElemOneOf) > 0 {
			out.Validators = []validator.Set{setvalidator.ValueStringsAre(stringvalidator.OneOf(a.ElemOneOf...))}
		}
		return out
	case NestedSet:
		return schema.SetNestedAttribute{
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			MarkdownDescription: a.Description,
			NestedObject:        schema.NestedAttributeObject{Attributes: a.nestedAttributes()},
			PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
		}
	case NestedList:
		return schema.ListNestedAttribute{
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			MarkdownDescription: a.Description,
			NestedObject:        schema.NestedAttributeObject{Attributes: a.nestedAttributes()},
			PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
		}
	default: // StringList, IntList
		out := schema.ListAttribute{
			ElementType:         a.elemType(),
			Required:            a.Required,
			Optional:            a.Optional,
			Computed:            computed,
			Sensitive:           a.Sensitive,
			MarkdownDescription: a.Description,
			DeprecationMessage:  a.Deprecated,
		}
		if a.Default != nil {
			out.Default = listdefault.StaticValue(types.ListValueMust(a.elemType(), []attr.Value{}))
		}
		if stable {
			out.PlanModifiers = append(out.PlanModifiers, listplanmodifier.UseStateForUnknown())
		}
		if a.ForceNew {
			out.PlanModifiers = append(out.PlanModifiers, listplanmodifier.RequiresReplace())
		}
		return out
	}
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	}
	return 0
}

func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}
