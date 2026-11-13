package fwresource

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const deviceTagDescription = ":meta:subcategory:Data Center Inventory Management (DCIM):This resource attaches a single tag to a device, similar in purpose to the `aws_ec2_tag` resource in the AWS provider.\n\nUse this when a device is created and managed outside of Terraform (e.g. via onboarding automation or the NetBox UI) but you still want Terraform to manage the presence of one or more specific tags on it, without taking ownership of the rest of the device.\n\nDo not use this on a device that is also managed by a `netbox_device` resource. Since that resource does not itself declare the tag, it will treat the tag added here as drift and try to remove it on every subsequent plan/apply; this is deterministic, not a race. Unlike the AWS provider, which has an `ignore_tags` provider setting for exactly this situation, this provider has no way to make a resource ignore a specific tag, so the two resources cannot currently coexist on the same device.\n\nThis resource performs a read-modify-write of the device's tag list. If the same device is targeted by more than one `netbox_device_tag` resource (or has its tags modified by anything else) at the same time, concurrent writes can race and one may silently drop the other's change."

const deviceAPIPath = "dcim/devices/"

// NewDeviceTag returns the constructor of the netbox_device_tag resource.
func NewDeviceTag() resource.Resource { return &deviceTag{} }

type deviceTag struct {
	state *providerstate.State
}

var (
	_ resource.Resource                = &deviceTag{}
	_ resource.ResourceWithConfigure   = &deviceTag{}
	_ resource.ResourceWithImportState = &deviceTag{}
)

func (r *deviceTag) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "netbox_device_tag"
}

func (r *deviceTag) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: deviceTagDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"device_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The ID of the device associated with the device tag.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"tag": schema.StringAttribute{
				Required:      true,
				Description:   "The tag of the device tag.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *deviceTag) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseDeviceTagID(id string) (int64, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid ID %q, expected format <device_id>:<tag>", id)
	}
	deviceID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid device ID in %q: %w", id, err)
	}
	return deviceID, parts[1], nil
}

func (r *deviceTag) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	deviceID, tag, err := parseDeviceTagID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), deviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), tag)...)
}

// deviceTags is the part of a device that this resource needs.
type deviceTags struct {
	ID   int64              `json:"id"`
	Tags []netboxapi.TagRef `json:"tags"`
}

// updateTags overwrites the tag list of a device. Only the tags are sent, everything else on the
// device is left alone.
func (r *deviceTag) updateTags(ctx context.Context, deviceID int64, tags []netboxapi.TagRef) error {
	body := map[string]any{"tags": tags}
	_, err := netboxapi.Patch[netboxapi.Ref](ctx, r.state.Rest, deviceAPIPath, deviceID, body)
	return err
}

func (r *deviceTag) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	var deviceID types.Int64
	var tagName types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("device_id"), &deviceID)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("tag"), &tagName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ref, err := r.state.TagRef(ctx, tagName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not find the tag", err.Error())
		return
	}
	device, err := netboxapi.Get[deviceTags](ctx, r.state.Rest, deviceAPIPath, deviceID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Could not read the device", err.Error())
		return
	}

	tags := device.Tags
	present := false
	for _, t := range tags {
		if t.Name == ref.Name {
			present = true
			break
		}
	}
	if !present {
		tags = append(tags, ref)
	}
	if err := r.updateTags(ctx, deviceID.ValueInt64(), tags); err != nil {
		resp.Diagnostics.AddError("Could not update the tags of the device", err.Error())
		return
	}

	resp.State.Raw = req.Plan.Raw
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%d:%s", deviceID.ValueInt64(), tagName.ValueString()))...)
}

func (r *deviceTag) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deviceID, tagName, err := parseDeviceTagID(id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}
	device, err := netboxapi.Get[deviceTags](ctx, r.state.Rest, deviceAPIPath, deviceID)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// the device is gone, so is the tag assignment
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Could not read the device", err.Error())
		return
	}
	for _, t := range device.Tags {
		if t.Name == tagName {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), deviceID)...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), tagName)...)
			return
		}
	}
	// the tag was removed from the device out of band
	resp.State.RemoveResource(ctx)
}

// Update is never called: both attributes force a new resource.
func (r *deviceTag) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {}

func (r *deviceTag) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if notConfigured(&resp.Diagnostics, r.state) {
		return
	}
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deviceID, tagName, err := parseDeviceTagID(id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}
	device, err := netboxapi.Get[deviceTags](ctx, r.state.Rest, deviceAPIPath, deviceID)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// the device is already gone, nothing left to clean up
			return
		}
		resp.Diagnostics.AddError("Could not read the device", err.Error())
		return
	}
	tags := []netboxapi.TagRef{}
	for _, t := range device.Tags {
		if t.Name != tagName {
			tags = append(tags, t)
		}
	}
	if err := r.updateTags(ctx, deviceID, tags); err != nil {
		resp.Diagnostics.AddError("Could not update the tags of the device", err.Error())
	}
}
