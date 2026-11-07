package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxAggregate() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxAggregateCreate,
		Read:   resourceNetboxAggregateRead,
		Update: resourceNetboxAggregateUpdate,
		Delete: resourceNetboxAggregateDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#aggregates):

> NetBox allows us to specify the portions of IP space that are interesting to us by defining aggregates. Typically, an aggregate will correspond to either an allocation of public (globally routable) IP space granted by a regional authority, or a private (internally-routable) designation.`,

		Schema: map[string]*schema.Schema{
			"prefix": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rir_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const aggregateAPIPath = "ipam/aggregates/"

type aggregateRequest struct {
	Prefix      string             `json:"prefix"`
	Description string             `json:"description,omitempty"`
	Tenant      *int64             `json:"tenant,omitempty"`
	Rir         *int64             `json:"rir,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type aggregateResponse struct {
	Prefix      string             `json:"prefix"`
	Description string             `json:"description"`
	Tenant      *netboxapi.Ref     `json:"tenant"`
	Rir         *netboxapi.Ref     `json:"rir"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func aggregateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*aggregateRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &aggregateRequest{
		Prefix:      d.Get("prefix").(string),
		Description: d.Get("description").(string),
		Tags:        tags,
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("rir_id"); ok {
		data.Rir = int64ToPtr(int64(v.(int)))
	}
	return data, nil
}

func resourceNetboxAggregateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := aggregateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, aggregateAPIPath, data, resourceNetboxAggregateRead)
}

func resourceNetboxAggregateRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, aggregateAPIPath, func(r *aggregateResponse) {
		d.Set("prefix", r.Prefix)
		d.Set("description", r.Description)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("rir_id", netboxapi.RefID(r.Rir))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxAggregateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := aggregateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, aggregateAPIPath, data, resourceNetboxAggregateRead)
}

func resourceNetboxAggregateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, aggregateAPIPath)
}
