package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The helpers in this file hold the create/read/update/delete plumbing that is
// the same for nearly every resource. A resource supplies the API path, the
// request body and a function that copies the response into the state.

// createObject creates an object and stores its ID, then runs read.
func createObject(d *schema.ResourceData, m interface{}, path string, body interface{}, read schema.ReadFunc) error {
	api := m.(*providerState)

	res, err := netboxapi.Create[netboxapi.Ref](context.Background(), api.rest, path, body)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.ID, 10))

	return read(d, m)
}

// readObject reads the object behind d.Id() and passes it to set. A missing
// object removes the resource from the state.
func readObject[T any](d *schema.ResourceData, m interface{}, path string, set func(*T)) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	res, err := netboxapi.Get[T](context.Background(), api.rest, path, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	set(res)
	return nil
}

// patchObject updates the object behind d.Id() with a PATCH, then runs read.
func patchObject(d *schema.ResourceData, m interface{}, path string, body interface{}, read schema.ReadFunc) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if _, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, path, id, body); err != nil {
		return err
	}

	return read(d, m)
}

// putObject updates the object behind d.Id() with a PUT, then runs read.
func putObject(d *schema.ResourceData, m interface{}, path string, body interface{}, read schema.ReadFunc) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if _, err := netboxapi.Update[netboxapi.Ref](context.Background(), api.rest, path, id, body); err != nil {
		return err
	}

	return read(d, m)
}

// deleteObject deletes the object behind d.Id(). An object that is already
// gone is not an error.
func deleteObject(d *schema.ResourceData, m interface{}, path string) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if err := netboxapi.Delete(context.Background(), api.rest, path, id); err != nil {
		if netboxapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	return nil
}

// slugFromResourceData returns the configured slug, or one generated from name.
func slugFromResourceData(d *schema.ResourceData, name string) string {
	if v, ok := d.GetOk("slug"); ok {
		return v.(string)
	}
	return getSlug(name)
}

// The *Ctx variants below are for resources that use the context-aware
// schema functions (CreateContext and friends).

// createObjectCtx is createObject for CreateContext resources.
func createObjectCtx(ctx context.Context, d *schema.ResourceData, m interface{}, path string, body interface{}, read schema.ReadContextFunc) diag.Diagnostics {
	api := m.(*providerState)

	res, err := netboxapi.Create[netboxapi.Ref](ctx, api.rest, path, body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(res.ID, 10))

	return read(ctx, d, m)
}

// readObjectCtx is readObject for ReadContext resources.
func readObjectCtx[T any](ctx context.Context, d *schema.ResourceData, m interface{}, path string, set func(*T)) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	res, err := netboxapi.Get[T](ctx, api.rest, path, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	set(res)
	return nil
}

// patchObjectCtx is patchObject for UpdateContext resources.
func patchObjectCtx(ctx context.Context, d *schema.ResourceData, m interface{}, path string, body interface{}, read schema.ReadContextFunc) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if _, err := netboxapi.Patch[netboxapi.Ref](ctx, api.rest, path, id, body); err != nil {
		return diag.FromErr(err)
	}

	return read(ctx, d, m)
}

// putObjectCtx is putObject for UpdateContext resources.
func putObjectCtx(ctx context.Context, d *schema.ResourceData, m interface{}, path string, body interface{}, read schema.ReadContextFunc) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if _, err := netboxapi.Update[netboxapi.Ref](ctx, api.rest, path, id, body); err != nil {
		return diag.FromErr(err)
	}

	return read(ctx, d, m)
}

// deleteObjectCtx is deleteObject for DeleteContext resources.
func deleteObjectCtx(ctx context.Context, d *schema.ResourceData, m interface{}, path string) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if err := netboxapi.Delete(ctx, api.rest, path, id); err != nil {
		if netboxapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

// nullableID returns the value to send for a reference field: null when the
// ID is 0, so that setting the attribute to nothing clears the reference.
func nullableID(id int) interface{} {
	if id == 0 {
		return nil
	}
	return int64(id)
}

// nullableString returns nil for an empty string, so that it is sent as null.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
