package netbox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// withContext turns a create, read, update or delete function that returns an
// error into the form the SDK wants for CreateContext, ReadContext,
// UpdateContext and DeleteContext. The function does not use the context yet,
// so it behaves as before.
func withContext(f func(d *schema.ResourceData, m interface{}) error) func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return diag.FromErr(f(d, m))
	}
}
