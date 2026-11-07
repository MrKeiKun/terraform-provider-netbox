package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxRir() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRirCreate,
		Read:   resourceNetboxRirRead,
		Update: resourceNetboxRirUpdate,
		Delete: resourceNetboxRirDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#regional-internet-registries-rirs):

> Regional Internet registries are responsible for the allocation of globally-routable address space. The five RIRs are ARIN, RIPE, APNIC, LACNIC, and AFRINIC. However, some address space has been set aside for internal use, such as defined in RFCs 1918 and 6598. NetBox considers these RFCs as a sort of RIR as well; that is, an authority which "owns" certain address space. There also exist lower-tier registries which serve particular geographic areas.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rirAPIPath = "ipam/rirs/"

type rirRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	IsPrivate   bool               `json:"is_private"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type rirResponse struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

func rirRequestFromResourceData(d *schema.ResourceData) *rirRequest {
	name := d.Get("name").(string)
	return &rirRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		IsPrivate:   d.Get("is_private").(bool),
		Tags:        []netboxapi.TagRef{},
	}
}

func resourceNetboxRirCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, rirAPIPath, rirRequestFromResourceData(d), resourceNetboxRirUpdate)
}

func resourceNetboxRirRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, rirAPIPath, func(rir *rirResponse) {
		d.Set("name", rir.Name)
		d.Set("slug", rir.Slug)
		d.Set("description", rir.Description)
		d.Set("is_private", rir.IsPrivate)
	})
}

func resourceNetboxRirUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, rirAPIPath, rirRequestFromResourceData(d), resourceNetboxRirRead)
}

func resourceNetboxRirDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, rirAPIPath)
}
