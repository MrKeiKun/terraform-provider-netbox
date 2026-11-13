package fwresource

import (
	"context"
	"fmt"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// isSet tells whether the configuration gives a value to the attribute. An empty set or list is the
// same as none.
func isSet(v attr.Value) bool {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return false
	}
	switch x := v.(type) {
	case types.Set:
		return len(x.Elements()) > 0
	case types.List:
		return len(x.Elements()) > 0
	case types.Map:
		return len(x.Elements()) > 0
	}
	return true
}

// checkVersion reports the object type and the attributes that the NetBox server is too old for.
// It does nothing when the version of the server is not known.
func (r *Resource) checkVersion(ctx context.Context, req resource.ModifyPlanRequest) diag.Diagnostics {
	var diags diag.Diagnostics
	if r.state == nil || r.state.Version == "" || (r.spec.SinceVersion == "" && len(r.spec.Since) == 0) {
		return diags
	}
	server, err := semver.NewVersion(r.state.Version)
	if err != nil {
		return diags
	}
	older := func(since string) bool {
		v, err := semver.NewVersion(since)
		return err == nil && server.LessThan(v)
	}
	if older(r.spec.SinceVersion) {
		diags.AddError("NetBox is too old", fmt.Sprintf("%s needs NetBox %s or later, and the server runs %s.", r.spec.TypeName, r.spec.SinceVersion, r.state.Version))
		return diags
	}
	if len(r.spec.Since) == 0 {
		return diags
	}
	config, d := r.spec.load(ctx, req.Config)
	diags.Append(d...)
	if d.HasError() {
		return diags
	}
	for name, since := range r.spec.Since {
		if isSet(config[name]) && older(since) {
			diags.AddAttributeError(path.Root(name), "NetBox is too old", fmt.Sprintf("The attribute %s of %s needs NetBox %s or later, and the server runs %s. NetBox would ignore it.", name, r.spec.TypeName, since, r.state.Version))
		}
	}
	return diags
}

// serverOlderThan tells whether the NetBox server is older than the version. It is false when the
// version of the server is not known.
func serverOlderThan(serverVersion, since string) bool {
	if serverVersion == "" || since == "" {
		return false
	}
	server, err := semver.NewVersion(serverVersion)
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(since)
	return err == nil && server.LessThan(v)
}

// MoveState lets a moved block turn the resources listed in Spec.MoveFrom into this one. They have the
// same attributes, so the state is taken over as it is.
func (r *Resource) MoveState(_ context.Context) []resource.StateMover {
	if len(r.spec.MoveFrom) == 0 {
		return nil
	}
	source := r.spec.buildSchema()
	return []resource.StateMover{{
		SourceSchema: &source,
		StateMover: func(_ context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			if req.SourceState == nil || !slices.Contains(r.spec.MoveFrom, req.SourceTypeName) {
				return
			}
			resp.TargetState.Raw = req.SourceState.Raw
		},
	}}
}

var _ resource.ResourceWithMoveState = &Resource{}
