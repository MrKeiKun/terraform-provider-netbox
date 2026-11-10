package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxFhrpGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxFhrpGroupCreate),
		ReadContext:   withContext(resourceNetboxFhrpGroupRead),
		UpdateContext: withContext(resourceNetboxFhrpGroupUpdate),
		DeleteContext: withContext(resourceNetboxFhrpGroupDelete),

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://netboxlabs.com/docs/netbox/models/ipam/fhrpgroup/):`,

		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Protocol",
			},
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Group ID",
			},
			"auth_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Auth type",
			},
			"auth_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Authentication key",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description",
			},
			"comments": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comments",
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the fhrp group.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const fhrpGroupAPIPath = "ipam/fhrp-groups/"

type fhrpGroupRequest struct {
	Protocol     string             `json:"protocol"`
	GroupId      *int64             `json:"group_id"`
	AuthType     string             `json:"auth_type,omitempty"`
	AuthKey      string             `json:"auth_key,omitempty"`
	Name         string             `json:"name,omitempty"`
	Description  string             `json:"description,omitempty"`
	Comments     string             `json:"comments,omitempty"`
	Owner        *int64             `json:"owner"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type fhrpGroupResponse struct {
	Protocol     string                 `json:"protocol"`
	GroupId      *int64                 `json:"group_id"`
	AuthType     string                 `json:"auth_type"`
	AuthKey      string                 `json:"auth_key"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Comments     string                 `json:"comments"`
	Owner        *netboxapi.Ref         `json:"owner"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func fhrpGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*fhrpGroupRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &fhrpGroupRequest{
		CustomFields: customFieldsRequest(d),
		Owner:        getOptionalInt(d, "owner_id"),
		Protocol:     d.Get("protocol").(string),
		GroupId:      int64ToPtr(int64(d.Get("group_id").(int))),
		AuthType:     d.Get("auth_type").(string),
		AuthKey:      d.Get("auth_key").(string),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Comments:     d.Get("comments").(string),
		Tags:         tags,
	}
	return data, nil
}

func resourceNetboxFhrpGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := fhrpGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, fhrpGroupAPIPath, data, resourceNetboxFhrpGroupRead)
}

func resourceNetboxFhrpGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, fhrpGroupAPIPath, func(r *fhrpGroupResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("protocol", r.Protocol)
		d.Set("group_id", r.GroupId)
		d.Set("auth_type", r.AuthType)
		d.Set("auth_key", r.AuthKey)
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxFhrpGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := fhrpGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, fhrpGroupAPIPath, data, resourceNetboxFhrpGroupRead)
}

func resourceNetboxFhrpGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, fhrpGroupAPIPath)
}
