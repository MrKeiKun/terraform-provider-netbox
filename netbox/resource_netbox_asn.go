package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxAsn() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxAsnCreate,
		Read:   resourceNetboxAsnRead,
		Update: resourceNetboxAsnUpdate,
		Delete: resourceNetboxAsnDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#asn):
> ASN is short for Autonomous System Number. This identifier is used in the BGP protocol to identify which "autonomous system" a particular prefix is originating and transiting through.
>
> The AS number model within NetBox allows you to model some of this real-world relationship.`,

		Schema: map[string]*schema.Schema{
			"asn": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Value for the AS Number record",
			},
			"rir_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID for the RIR for the AS Number record",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description field for the AS Number record",
			},
			"comments": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comments field for the AS Number record",
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const asnAPIPath = "ipam/asns/"

type asnRequest struct {
	Asn         *int64             `json:"asn"`
	Rir         *int64             `json:"rir"`
	Description string             `json:"description,omitempty"`
	Comments    string             `json:"comments,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type asnResponse struct {
	Asn         *int64             `json:"asn"`
	Rir         *netboxapi.Ref     `json:"rir"`
	Description string             `json:"description"`
	Comments    string             `json:"comments"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func asnRequestFromResourceData(api *providerState, d *schema.ResourceData) (*asnRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &asnRequest{
		Asn:         int64ToPtr(int64(d.Get("asn").(int))),
		Rir:         int64ToPtr(int64(d.Get("rir_id").(int))),
		Description: d.Get("description").(string),
		Comments:    d.Get("comments").(string),
		Tags:        tags,
	}
	return data, nil
}

func resourceNetboxAsnCreate(d *schema.ResourceData, m interface{}) error {
	data, err := asnRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, asnAPIPath, data, resourceNetboxAsnRead)
}

func resourceNetboxAsnRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, asnAPIPath, func(r *asnResponse) {
		d.Set("asn", r.Asn)
		d.Set("rir_id", netboxapi.RefID(r.Rir))
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxAsnUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := asnRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, asnAPIPath, data, resourceNetboxAsnRead)
}

func resourceNetboxAsnDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, asnAPIPath)
}
