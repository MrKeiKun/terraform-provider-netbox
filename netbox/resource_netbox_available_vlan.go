package netbox

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxAvailableVLAN() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxAvailableVLANCreate,
		Read:          resourceNetboxAvailableVLANRead,
		Update:        resourceNetboxAvailableVLANUpdate,
		Delete:        resourceNetboxAvailableVLANDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):Per [the docs](https://netbox.readthedocs.io/en/stable/models/ipam/vlan/):

> A VLAN represents an isolated Layer 2 domain identified by a numeric ID (1–4094). VLANs may be assigned to specific sites or marked as global.
> Optionally, they can be organized within VLAN groups to define scope and enforce uniqueness.
>
> Each VLAN can also be assigned an operational status and a functional role. Statuses are hard-coded in NetBox and include the following:
> * Active
> * Reserved
> * Deprecated

This resource will retrieve the next available VLAN ID from a given VLAN group (specified by ID).`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the available vlan.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the available vlan.",
				Optional:    true,
			},
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the available vlan.",
				Optional:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the available vlan.",
				Optional:    true,
			},
			"site_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site associated with the available vlan.",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the available vlan.",
				Optional:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the available vlan.",
				Optional:    true,
			},
			"vid": {
				Type:        schema.TypeInt,
				Description: "The VLAN ID (1-4094).",
				Computed:    true, // it's auto-assigned by NetBox, not user-supplied
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the available vlan.",
				Computed:    true,
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

type availableVLANCreateRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Tenant      *int64             `json:"tenant,omitempty"`
	Site        *int64             `json:"site,omitempty"`
	Role        *int64             `json:"role,omitempty"`
	Status      string             `json:"status,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type availableVLANUpdateRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Tenant      *int64             `json:"tenant,omitempty"`
	Site        *int64             `json:"site,omitempty"`
	Group       *int64             `json:"group,omitempty"`
	Role        *int64             `json:"role,omitempty"`
	Status      string             `json:"status,omitempty"`
	Vid         *int64             `json:"vid"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type availableVLANResponse struct {
	ID          int64              `json:"id"`
	Vid         int64              `json:"vid"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Comments    string             `json:"comments"`
	Status      *netboxapi.Choice  `json:"status"`
	Tenant      *netboxapi.Ref     `json:"tenant"`
	Site        *netboxapi.Ref     `json:"site"`
	Group       *netboxapi.Ref     `json:"group"`
	Role        *netboxapi.Ref     `json:"role"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func resourceNetboxAvailableVLANCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	groupID := int64(d.Get("group_id").(int))

	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsKey))
	if err != nil {
		return diag.FromErr(err)
	}

	data := &availableVLANCreateRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", false),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Site:        getOptionalInt(d, "site_id"),
		Role:        getOptionalInt(d, "role_id"),
		Status:      d.Get("status").(string),
		Tags:        tags,
	}

	// NetBox picks the next free VLAN ID of the group and returns the new VLAN.
	path := fmt.Sprintf("ipam/vlan-groups/%d/available-vlans/", groupID)

	// Only the allocation itself is serialized and retried. Everything after it
	// operates on the VLAN we just got, which nothing else is competing for.
	unlock, err := api.lockAllocation(ctx, allocationLockKeyVLANGroup(groupID))
	if err != nil {
		return diag.FromErr(err)
	}
	defer unlock()
	var vlan *availableVLANResponse
	err = retryAllocation(ctx, func() error {
		var err error
		vlan, err = netboxapi.Request[availableVLANResponse](ctx, api.rest, http.MethodPost, path, nil, data)
		return err
	})
	unlock()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(vlan.ID, 10))
	d.Set("vid", vlan.Vid)
	d.Set("name", vlan.Name)
	if vlan.Group != nil {
		d.Set("group_id", vlan.Group.ID)
	}

	return diag.FromErr(resourceNetboxAvailableVLANRead(d, m))
}

func resourceNetboxAvailableVLANRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, "ipam/vlans/", func(vlan *availableVLANResponse) {
		// Required fields
		d.Set("vid", vlan.Vid)
		d.Set("name", vlan.Name)

		// Optional fields
		d.Set("description", vlan.Description)
		d.Set("comments", vlan.Comments)
		d.Set("status", netboxapi.ChoiceValue(vlan.Status))
		d.Set("tenant_id", netboxapi.RefID(vlan.Tenant))
		d.Set("site_id", netboxapi.RefID(vlan.Site))
		d.Set("group_id", netboxapi.RefID(vlan.Group))
		d.Set("role_id", netboxapi.RefID(vlan.Role))
		api.readTagRefs(d, vlan.Tags)
	})
}

func resourceNetboxAvailableVLANUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsKey))
	if err != nil {
		return err
	}

	data := &availableVLANUpdateRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", false),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Site:        getOptionalInt(d, "site_id"),
		Group:       getOptionalInt(d, "group_id"),
		Role:        getOptionalInt(d, "role_id"),
		Status:      d.Get("status").(string),
		Vid:         int64ToPtr(int64(d.Get("vid").(int))),
		Tags:        tags,
	}

	return putObject(d, m, "ipam/vlans/", data, resourceNetboxAvailableVLANRead)
}

func resourceNetboxAvailableVLANDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, "ipam/vlans/")
}
