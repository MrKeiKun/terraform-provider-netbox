package fwresource

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AssignSpec describes a resource that points one reference field of an
// existing object (for example the primary IP of a device) at another object.
// The resource ID is the ID of the object that is changed.
type AssignSpec struct {
	TypeName          string
	Description       string
	Path              string
	Target            string // attribute holding the ID of the changed object
	TargetDescription string
	Ref               string // attribute holding the ID of the referenced object
	RefDescription    string
	// VersionDescription enables the ip_address_version attribute. The
	// field is then primary_ip4 or primary_ip6.
	VersionDescription string
	// Field is the API field that is set when there is no version.
	Field string
	// KeepVLANs sends the VLANs of the object along, because NetBox drops
	// them on every save of an interface that is not in tagged mode.
	KeepVLANs bool
}

// NewAssign returns the constructor of an assignment resource.
func NewAssign(spec AssignSpec) func() resource.Resource {
	return func() resource.Resource { return &assignResource{spec: spec} }
}

type assignResource struct {
	spec  AssignSpec
	state *providerstate.State
}

var (
	_ resource.Resource                = &assignResource{}
	_ resource.ResourceWithConfigure   = &assignResource{}
	_ resource.ResourceWithImportState = &assignResource{}
)

const versionAttr = "ip_address_version"

func (r *assignResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.spec.TypeName
}

func (r *assignResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		r.spec.Target: schema.Int64Attribute{
			Required:      true,
			Description:   r.spec.TargetDescription,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
		},
		r.spec.Ref: schema.Int64Attribute{
			Required:    true,
			Description: r.spec.RefDescription,
		},
	}
	if r.spec.VersionDescription != "" {
		attrs[versionAttr] = schema.Int64Attribute{
			Optional:      true,
			Computed:      true,
			Default:       int64default.StaticInt64(4),
			Description:   r.spec.VersionDescription,
			Validators:    []validator.Int64{int64validator.OneOf(4, 6)},
			PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
		}
	}
	resp.Schema = schema.Schema{Description: r.spec.Description, Attributes: attrs}
}

func (r *assignResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *assignResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("%q is not a number", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(r.spec.Target), id)...)
	// the reference (and the version) are filled by the read after the import
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(r.spec.Ref), int64(0))...)
	if r.spec.VersionDescription != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(versionAttr), int64(0))...)
	}
}

func (r *assignResource) field(version int64) string {
	if r.spec.VersionDescription == "" {
		return r.spec.Field
	}
	if version == 6 {
		return "primary_ip6"
	}
	return "primary_ip4"
}

// patch points the field at ref (nil clears it).
func (r *assignResource) patch(ctx context.Context, target int64, version int64, ref *int64) error {
	body := map[string]any{}
	if ref == nil {
		body[r.field(version)] = nil
	} else {
		body[r.field(version)] = *ref
	}
	if r.spec.KeepVLANs {
		cur, err := netboxapi.Get[map[string]any](ctx, r.state.Rest, r.spec.Path, target)
		if err != nil {
			return err
		}
		tagged := []int64{}
		if list, ok := (*cur)["tagged_vlans"].([]any); ok {
			for _, v := range list {
				if id, ok := refID(v); ok {
					tagged = append(tagged, id)
				}
			}
		}
		body["tagged_vlans"] = tagged
		if id, ok := refID((*cur)["untagged_vlan"]); ok {
			body["untagged_vlan"] = id
		}
	}
	_, err := netboxapi.Request[map[string]any](ctx, r.state.Rest, http.MethodPatch, fmt.Sprintf("%s%d/", r.spec.Path, target), nil, body)
	return err
}

func refID(v any) (int64, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return 0, false
	}
	f, ok := m["id"].(float64)
	return int64(f), ok
}

func (r *assignResource) values(ctx context.Context, g getter) (target, ref, version int64, err error) {
	var t, rf, v types.Int64
	d := g.GetAttribute(ctx, path.Root(r.spec.Target), &t)
	d.Append(g.GetAttribute(ctx, path.Root(r.spec.Ref), &rf)...)
	if r.spec.VersionDescription != "" {
		d.Append(g.GetAttribute(ctx, path.Root(versionAttr), &v)...)
	}
	if d.HasError() {
		return 0, 0, 0, fmt.Errorf("%s", d.Errors()[0].Detail())
	}
	return t.ValueInt64(), rf.ValueInt64(), v.ValueInt64(), nil
}

func (r *assignResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	target, ref, version, err := r.values(ctx, req.Plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}
	if err := r.patch(ctx, target, version, &ref); err != nil {
		resp.Diagnostics.AddError("Could not assign "+r.spec.Ref, err.Error())
		return
	}
	resp.State.Raw = req.Plan.Raw
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.FormatInt(target, 10))...)
}

func (r *assignResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	target, ref, version, err := r.values(ctx, req.Plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}
	if err := r.patch(ctx, target, version, &ref); err != nil {
		resp.Diagnostics.AddError("Could not assign "+r.spec.Ref, err.Error())
		return
	}
	resp.State.Raw = req.Plan.Raw
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.FormatInt(target, 10))...)
}

func (r *assignResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	target, _, version, err := r.values(ctx, req.State)
	if err != nil {
		resp.Diagnostics.AddError("Invalid state", err.Error())
		return
	}
	obj, err := netboxapi.Get[map[string]any](ctx, r.state.Rest, r.spec.Path, target)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Could not read "+r.spec.TypeName, err.Error())
		return
	}
	if r.spec.VersionDescription != "" && version == 0 {
		// On import the version is unknown; infer it from whichever IP is set.
		version = 4
		if _, ok := refID((*obj)["primary_ip4"]); !ok {
			if _, ok := refID((*obj)["primary_ip6"]); ok {
				version = 6
			}
		}
	}
	id, ok := refID((*obj)[r.field(version)])
	if !ok {
		// the object exists but nothing is assigned: the assignment is gone
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(r.spec.Ref), id)...)
	if r.spec.VersionDescription != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(versionAttr), version)...)
	}
}

func (r *assignResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	target, _, version, err := r.values(ctx, req.State)
	if err != nil {
		resp.Diagnostics.AddError("Invalid state", err.Error())
		return
	}
	if err := r.patch(ctx, target, version, nil); err != nil {
		if netboxapi.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Could not clear "+r.spec.Ref, err.Error())
	}
}
