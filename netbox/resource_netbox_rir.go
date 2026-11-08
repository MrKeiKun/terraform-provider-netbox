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
				Type:        schema.TypeString,
				Description: "The name of the rir.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the rir.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the rir.",
				Optional:    true,
			},
			"is_private": {
				Type:        schema.TypeBool,
				Description: "Whether the rir is private.",
				Optional:    true,
				Default:     false,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the rir.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the rir.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rirAPIPath = "ipam/rirs/"

type rirRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	IsPrivate    bool               `json:"is_private"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type rirResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	IsPrivate    bool                   `json:"is_private"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func rirRequestFromResourceData(d *schema.ResourceData) *rirRequest {
	name := d.Get("name").(string)
	return &rirRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Description:  getOptionalStr(d, "description", true),
		IsPrivate:    d.Get("is_private").(bool),
		Tags:         []netboxapi.TagRef{},
	}
}

func resourceNetboxRirCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, rirAPIPath, rirRequestFromResourceData(d), resourceNetboxRirUpdate)
}

func resourceNetboxRirRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, rirAPIPath, func(rir *rirResponse) {
		d.Set(customFieldsKey, getCustomFields(rir.CustomFields))
		d.Set("comments", rir.Comments)
		d.Set("owner_id", netboxapi.RefID(rir.Owner))
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
