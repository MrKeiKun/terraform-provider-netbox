package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxASNRange() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxASNRangeCreate),
		ReadContext:   withContext(resourceNetboxASNRangeRead),
		UpdateContext: withContext(resourceNetboxASNRangeUpdate),
		DeleteContext: withContext(resourceNetboxASNRangeDelete),

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/models/ipam/asnrange/):

> An ASN range is a block of AS numbers that belongs to a RIR.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the asn range.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the asn range.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"rir_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the rir associated with the asn range.",
				Required:    true,
			},
			"start": {
				Type:         schema.TypeInt,
				Description:  "The start of the asn range.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"end": {
				Type:         schema.TypeInt,
				Description:  "The end of the asn range.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the asn range.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the asn range.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the asn range.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the asn range.",
				Optional:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const aSNRangeAPIPath = "ipam/asn-ranges/"

type aSNRangeRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Rir          *int64             `json:"rir"`
	Start        int64              `json:"start"`
	End          int64              `json:"end"`
	Tenant       *int64             `json:"tenant"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type aSNRangeResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Rir          *netboxapi.Ref         `json:"rir"`
	Start        *int64                 `json:"start"`
	End          *int64                 `json:"end"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func aSNRangeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*aSNRangeRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &aSNRangeRequest{
		Name:        d.Get("name").(string),
		Slug:        slugFromResourceData(d, name),
		Rir:         int64ToPtr(int64(d.Get("rir_id").(int))),
		Start:       int64(d.Get("start").(int)),
		End:         int64(d.Get("end").(int)),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxASNRangeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := aSNRangeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, aSNRangeAPIPath, data, resourceNetboxASNRangeRead)
}

func resourceNetboxASNRangeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, aSNRangeAPIPath, func(r *aSNRangeResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("rir_id", netboxapi.RefID(r.Rir))
		d.Set("start", r.Start)
		d.Set("end", r.End)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxASNRangeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := aSNRangeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, aSNRangeAPIPath, data, resourceNetboxASNRangeRead)
}

func resourceNetboxASNRangeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, aSNRangeAPIPath)
}
