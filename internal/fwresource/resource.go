package fwresource

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// New returns the constructor of a resource described by spec.
func New(spec *Spec) func() resource.Resource {
	spec = spec.withZeroDefaults()
	return func() resource.Resource {
		return &Resource{spec: spec}
	}
}

// Resource is a NetBox object managed through the REST API.
type Resource struct {
	spec  *Spec
	state *providerstate.State
}

var (
	_ resource.Resource                     = &Resource{}
	_ resource.ResourceWithConfigure        = &Resource{}
	_ resource.ResourceWithImportState      = &Resource{}
	_ resource.ResourceWithModifyPlan       = &Resource{}
	_ resource.ResourceWithUpgradeState     = &Resource{}
	_ resource.ResourceWithConfigValidators = &Resource{}
	_ resource.ConfigValidator              = requiredWith{}
)

func (r *Resource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.spec.TypeName
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.spec.buildSchema()
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	state, ok := req.ProviderData.(*providerstate.State)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *providerstate.State, got %T", req.ProviderData))
		return
	}
	r.state = state
}

// UpgradeState upgrades the state of older versions of the schema.
func (r *Resource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	out := map[int64]resource.StateUpgrader{}
	for version, fn := range r.spec.Upgraders {
		fn := fn
		out[version] = resource.StateUpgrader{
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				if req.RawState == nil || req.RawState.JSON == nil {
					resp.Diagnostics.AddError("Cannot upgrade the state", "the state has no JSON representation")
					return
				}
				var state map[string]any
				if err := json.Unmarshal(req.RawState.JSON, &state); err != nil {
					resp.Diagnostics.AddError("Cannot upgrade the state", err.Error())
					return
				}
				upgraded, err := fn(ctx, state)
				if err != nil {
					resp.Diagnostics.AddError("Cannot upgrade the state", err.Error())
					return
				}
				b, err := json.Marshal(upgraded)
				if err != nil {
					resp.Diagnostics.AddError("Cannot upgrade the state", err.Error())
					return
				}
				resp.DynamicValue = &tfprotov6.DynamicValue{JSON: b}
			},
		}
	}
	return out
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	paths := func(names []string) []path.Expression {
		out := make([]path.Expression, len(names))
		for i, n := range names {
			out[i] = path.MatchRoot(n)
		}
		return out
	}
	var v []resource.ConfigValidator
	for _, g := range r.spec.ExactlyOneOf {
		v = append(v, resourcevalidator.ExactlyOneOf(paths(g)...))
	}
	for _, g := range r.spec.AtLeastOneOf {
		v = append(v, resourcevalidator.AtLeastOneOf(paths(g)...))
	}
	for _, g := range r.spec.Conflicts {
		v = append(v, resourcevalidator.Conflicting(paths(g)...))
	}
	for _, g := range r.spec.RequiredWith {
		if len(g) == 2 {
			v = append(v, requiredWith{attr: g[0], other: g[1]})
		}
	}
	return v
}

type getter interface {
	GetAttribute(ctx context.Context, p path.Path, target any) diag.Diagnostics
}

type setter interface {
	SetAttribute(ctx context.Context, p path.Path, val any) diag.Diagnostics
}

// emptyValue returns a null value of the type of the attribute.
func (a Attr) target() any {
	switch a.Kind {
	case Int:
		return &types.Int64{}
	case Float:
		return &types.Float64{}
	case Bool:
		return &types.Bool{}
	case StringSet, IntSet, NestedSet:
		return &types.Set{}
	case StringList, IntList, NestedList, IntLists, StringLists:
		return &types.List{}
	case StringMap:
		return &types.Map{}
	}
	return &types.String{}
}

func deref(t any) attr.Value {
	switch v := t.(type) {
	case *types.String:
		return *v
	case *types.Int64:
		return *v
	case *types.Float64:
		return *v
	case *types.Bool:
		return *v
	case *types.Set:
		return *v
	case *types.List:
		return *v
	case *types.Map:
		return *v
	}
	return nil
}

// load reads the values of all attributes.
func (s *Spec) load(ctx context.Context, g getter) (Values, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := Values{}
	get := func(name string, target any) {
		diags.Append(g.GetAttribute(ctx, path.Root(name), target)...)
		out[name] = deref(target)
	}
	get("id", &types.String{})
	for _, a := range s.Attrs {
		get(a.Name, a.target())
	}
	if s.Tags {
		get(tagsKey, &types.Set{})
		get(tagsAllKey, &types.Set{})
	}
	if s.CustomFields {
		get(customFieldsKey, &types.Map{})
	}
	return out, diags
}

// store writes the values to the state.
func (s *Spec) store(ctx context.Context, st setter, vals Values) diag.Diagnostics {
	var diags diag.Diagnostics
	for name, v := range vals {
		diags.Append(st.SetAttribute(ctx, path.Root(name), v)...)
	}
	return diags
}

func stringSetValues(v attr.Value) []string {
	set, ok := v.(types.Set)
	if !ok || set.IsNull() || set.IsUnknown() {
		return nil
	}
	var out []string
	for _, e := range set.Elements() {
		if s, ok := e.(types.String); ok {
			out = append(out, s.ValueString())
		}
	}
	return out
}

func stringMapValues(v attr.Value) (map[string]string, bool) {
	m, ok := v.(types.Map)
	if !ok || m.IsNull() || m.IsUnknown() {
		return nil, false
	}
	out := make(map[string]string, len(m.Elements()))
	for k, e := range m.Elements() {
		if s, ok := e.(types.String); ok && !s.IsNull() {
			out[k] = s.ValueString()
		}
	}
	return out, true
}

var (
	slugSpecial    = regexp.MustCompile(`[^\w\s-]`)
	slugWhitespace = regexp.MustCompile(`[\s-]+`)
)

// Slug turns a name into a slug the way NetBox does.
func Slug(name string) string {
	result := slugSpecial.ReplaceAllString(name, "")
	result = slugWhitespace.ReplaceAllString(result, "-")
	return strings.ToLower(strings.Trim(result, "-"))
}

// body builds the request body from the planned values. prior is nil on create.
func (r *Resource) body(ctx context.Context, plan, prior Values) (map[string]any, error) {
	body := map[string]any{}
	virtual := r.spec.virtualAttrs()
	for _, a := range r.spec.Attrs {
		if virtual[a.Name] || a.NoSend || a.Local || (a.Computed && !a.Optional && !a.Required && a.Default == nil) {
			continue
		}
		var priorValue attr.Value
		if prior != nil {
			priorValue = prior[a.Name]
			if priorValue == nil {
				priorValue = types.StringNull()
			}
		}
		v, ok, err := a.toJSON(plan[a.Name], priorValue)
		if err != nil {
			return nil, err
		}
		if ok {
			body[a.key()] = v
		}
	}

	r.spec.scopeBody(plan, body)

	if r.spec.SlugFrom != "" {
		if slug, ok := plan["slug"].(types.String); ok && (slug.IsNull() || slug.IsUnknown()) {
			if name, ok := plan[r.spec.SlugFrom].(types.String); ok && !name.IsNull() && !name.IsUnknown() {
				body["slug"] = Slug(name.ValueString())
			}
		}
	}

	if r.spec.Tags {
		names := stringSetValues(plan[tagsAllKey])
		refs := make([]netboxapi.TagRef, 0, len(names))
		for _, n := range names {
			ref, err := r.state.TagRef(ctx, n)
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)
		}
		body["tags"] = refs
	}

	if r.spec.CustomFields {
		if cf, ok := stringMapValues(plan[customFieldsKey]); ok && len(cf) > 0 {
			m := make(map[string]any, len(cf))
			for k, v := range cf {
				m[k] = v
			}
			body["custom_fields"] = m
		}
	}

	if r.spec.ModifyBody != nil {
		if err := r.spec.ModifyBody(ctx, plan, prior, body); err != nil {
			return nil, err
		}
	}
	return body, nil
}

// stateFromObject converts an API response into attribute values. prior holds the values of the
// plan or state; adopt tells to take over values of attributes that were not set, as when reading
// the current state of the object.
func (r *Resource) stateFromObject(ctx context.Context, obj map[string]any, prior Values, adopt bool) (Values, error) {
	out := Values{}

	id, ok := intFromJSON(obj["id"])
	if !ok {
		return nil, fmt.Errorf("the response has no id")
	}
	out["id"] = types.StringValue(strconv.FormatInt(id, 10))

	virtual := r.spec.virtualAttrs()
	for _, a := range r.spec.Attrs {
		p := prior[a.Name]
		switch {
		case virtual[a.Name]:
			continue
		case a.WriteOnly || a.Local:
			if p == nil || p.IsUnknown() {
				out[a.Name] = nullOf(a)
			} else {
				out[a.Name] = p
			}
		case !adopt && !a.Computed && a.Default == nil && (p == nil || p.IsNull()):
			out[a.Name] = nullOf(a)
		default:
			out[a.Name] = a.fromJSON(a.pick(obj), p)
		}
	}

	r.spec.scopeState(obj, prior, out)

	if r.spec.Tags {
		var names []string
		if list, ok := obj["tags"].([]any); ok {
			for _, t := range list {
				if m, ok := t.(map[string]any); ok {
					if n, ok := m["name"].(string); ok {
						names = append(names, n)
					}
				}
			}
		}
		configured := map[string]bool{}
		for _, n := range stringSetValues(prior[tagsKey]) {
			configured[n] = true
		}
		all := make([]attr.Value, 0, len(names))
		own := make([]attr.Value, 0, len(names))
		for _, n := range names {
			all = append(all, types.StringValue(n))
			if !r.state.DefaultTags[n] || configured[n] {
				own = append(own, types.StringValue(n))
			}
		}
		out[tagsAllKey], _ = types.SetValue(types.StringType, all)
		if len(own) == 0 && (prior[tagsKey] == nil || prior[tagsKey].IsNull()) {
			out[tagsKey] = types.SetNull(types.StringType)
		} else {
			out[tagsKey], _ = types.SetValue(types.StringType, own)
		}
	}

	if r.spec.CustomFields {
		only, set := stringMapValues(prior[customFieldsKey])
		fields := customFieldsFromObject(obj["custom_fields"], only, !set, !adopt)
		elems := make(map[string]attr.Value, len(fields))
		for k, v := range fields {
			elems[k] = types.StringValue(v)
		}
		m, _ := types.MapValue(types.StringType, elems)
		out[customFieldsKey] = m
	}

	if r.spec.ModifyState != nil {
		if err := r.spec.ModifyState(ctx, obj, prior, out); err != nil {
			return nil, err
		}
	}
	if r.spec.Finish != nil {
		if err := r.spec.Finish(ctx, r.state, obj, prior, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func nullOf(a Attr) attr.Value {
	switch a.Kind {
	case Int:
		return types.Int64Null()
	case Float:
		return types.Float64Null()
	case Bool:
		return types.BoolNull()
	case StringSet, IntSet, StringList, IntList, NestedSet, NestedList, IntLists, StringLists:
		return a.nullCollection()
	case StringMap:
		return types.MapNull(types.StringType)
	}
	return types.StringNull()
}

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.state == nil {
		return
	}
	if r.spec.ModifyPlan != nil {
		config, diags := r.spec.load(ctx, req.Config)
		resp.Diagnostics.Append(diags...)
		plan, diags := r.spec.load(ctx, req.Plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for name, v := range r.spec.ModifyPlan(config, plan) {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(name), v)...)
		}
	}
	if !r.spec.Tags {
		return
	}
	var tags types.Set
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(tagsKey), &tags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	unknownElem := false
	for _, e := range tags.Elements() {
		if e.IsUnknown() {
			unknownElem = true
		}
	}
	if tags.IsUnknown() || unknownElem {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(tagsAllKey), types.SetUnknown(types.StringType))...)
		return
	}
	seen := map[string]bool{}
	var elems []attr.Value
	add := func(n string) {
		if !seen[n] {
			seen[n] = true
			elems = append(elems, types.StringValue(n))
		}
	}
	for _, n := range stringSetValues(tags) {
		add(n)
	}
	if !r.spec.NoDefaultTags {
		for n := range r.state.DefaultTags {
			add(n)
		}
	}
	all, _ := types.SetValue(types.StringType, elems)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(tagsAllKey), all)...)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan, diags := r.spec.load(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.body(ctx, r.configured(ctx, req.Config, plan), nil)
	if err != nil {
		resp.Diagnostics.AddError("Could not build the request", err.Error())
		return
	}

	var obj *map[string]any
	if r.spec.Allocate != nil {
		obj, err = r.allocate(ctx, req.Config, plan, body)
	} else {
		obj, err = netboxapi.Create[map[string]any](ctx, r.state.Rest, r.spec.Path, body)
	}
	if err != nil {
		resp.Diagnostics.AddError("Could not create "+r.spec.TypeName, err.Error())
		return
	}

	if obj, err = r.afterWrite(ctx, obj, plan, nil); err != nil {
		resp.Diagnostics.AddError("Could not create "+r.spec.TypeName, err.Error())
		return
	}

	vals, err := r.stateFromObject(ctx, *obj, plan, false)
	if err != nil {
		resp.Diagnostics.AddError("Could not read the created object", err.Error())
		return
	}
	resp.Diagnostics.Append(r.spec.store(ctx, &resp.State, vals)...)
}

// afterWrite runs the hook that follows a write and reads the object again when it changed it.
func (r *Resource) afterWrite(ctx context.Context, obj *map[string]any, plan, prior Values) (*map[string]any, error) {
	if r.spec.AfterWrite == nil {
		return obj, nil
	}
	id, ok := intFromJSON((*obj)["id"])
	if !ok {
		return obj, nil
	}
	changed, err := r.spec.AfterWrite(ctx, r.state, id, plan, prior)
	if err != nil || !changed {
		return obj, err
	}
	return netboxapi.Get[map[string]any](ctx, r.state.Rest, r.spec.Path, id)
}

// allocate lets NetBox pick the object and then sets all attributes with an update.
func (r *Resource) allocate(ctx context.Context, cfg getter, plan Values, body map[string]any) (*map[string]any, error) {
	picked, err := r.spec.Allocate(ctx, r.state, plan, body)
	if err != nil {
		return nil, err
	}
	id, ok := intFromJSON(picked["id"])
	if !ok {
		return nil, fmt.Errorf("the response has no id")
	}
	prior, err := r.stateFromObject(ctx, picked, plan, true)
	if err != nil {
		return nil, err
	}
	update, err := r.body(ctx, r.configured(ctx, cfg, plan), prior)
	if err != nil {
		return nil, err
	}
	if r.spec.Put {
		return netboxapi.Update[map[string]any](ctx, r.state.Rest, r.spec.Path, id, update)
	}
	return netboxapi.Patch[map[string]any](ctx, r.state.Rest, r.spec.Path, id, update)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	prior, diags := r.spec.load(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(prior["id"].(types.String).ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("%q is not a number", prior["id"]))
		return
	}

	obj, err := netboxapi.Get[map[string]any](ctx, r.state.Rest, r.spec.Path, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Could not read "+r.spec.TypeName, err.Error())
		return
	}

	vals, err := r.stateFromObject(ctx, *obj, prior, true)
	if err != nil {
		resp.Diagnostics.AddError("Could not read the object", err.Error())
		return
	}
	resp.Diagnostics.Append(r.spec.store(ctx, &resp.State, vals)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan, diags := r.spec.load(ctx, req.Plan)
	resp.Diagnostics.Append(diags...)
	prior, diags := r.spec.load(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(prior["id"].(types.String).ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	body, err := r.body(ctx, r.configured(ctx, req.Config, plan), prior)
	if err != nil {
		resp.Diagnostics.AddError("Could not build the request", err.Error())
		return
	}

	var obj *map[string]any
	if r.spec.Put {
		obj, err = netboxapi.Update[map[string]any](ctx, r.state.Rest, r.spec.Path, id, body)
	} else {
		obj, err = netboxapi.Patch[map[string]any](ctx, r.state.Rest, r.spec.Path, id, body)
	}
	if err != nil {
		resp.Diagnostics.AddError("Could not update "+r.spec.TypeName, err.Error())
		return
	}

	if obj, err = r.afterWrite(ctx, obj, plan, prior); err != nil {
		resp.Diagnostics.AddError("Could not update "+r.spec.TypeName, err.Error())
		return
	}

	vals, err := r.stateFromObject(ctx, *obj, plan, false)
	if err != nil {
		resp.Diagnostics.AddError("Could not read the updated object", err.Error())
		return
	}
	resp.Diagnostics.Append(r.spec.store(ctx, &resp.State, vals)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var idStr types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &idStr)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(idStr.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}
	if r.spec.BeforeDelete != nil {
		state, diags := r.spec.load(ctx, req.State)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := r.spec.BeforeDelete(ctx, r.state, id, state); err != nil {
			resp.Diagnostics.AddError("Could not delete "+r.spec.TypeName, err.Error())
			return
		}
	}
	if err := netboxapi.Delete(ctx, r.state.Rest, r.spec.Path, id); err != nil && !netboxapi.IsNotFound(err) {
		resp.Diagnostics.AddError("Could not delete "+r.spec.TypeName, err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.spec.ImportID != nil {
		id, attrs, err := r.spec.ImportID(req.ID)
		if err != nil {
			resp.Diagnostics.AddError("Invalid import ID", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		for k, v := range attrs {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(k), v)...)
		}
		return
	}
	if _, err := strconv.ParseInt(req.ID, 10, 64); err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("%q is not a number", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// requiredWith is the validator for an attribute that needs another one: when attr is set in the
// configuration, other has to be set too.
type requiredWith struct {
	attr, other string
}

func (v requiredWith) Description(context.Context) string {
	return fmt.Sprintf("%s requires %s", v.attr, v.other)
}

func (v requiredWith) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requiredWith) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var obj map[string]tftypes.Value
	if err := req.Config.Raw.As(&obj); err != nil {
		return
	}
	a, b := obj[v.attr], obj[v.other]
	if a.IsNull() || !a.IsKnown() {
		return
	}
	if b.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root(v.attr), "Missing attribute configuration", fmt.Sprintf("Attribute %q must be specified when %q is specified", v.other, v.attr))
	}
}

// configured returns the planned values, except that custom fields are the ones in the
// configuration. Custom fields that only exist in the state were not asked for, and the field
// may not exist any more.
func (r *Resource) configured(ctx context.Context, cfg getter, plan Values) Values {
	if !r.spec.CustomFields {
		return plan
	}
	out := make(Values, len(plan))
	for k, v := range plan {
		out[k] = v
	}
	var cf types.Map
	if diags := cfg.GetAttribute(ctx, path.Root(customFieldsKey), &cf); diags.HasError() {
		return plan
	}
	out[customFieldsKey] = cf
	return out
}
