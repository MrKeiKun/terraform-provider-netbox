package netbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxAvailablePrefix() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxAvailablePrefixCreate,
		Read:   resourceNetboxPrefixRead,
		Update: resourceNetboxPrefixUpdate,
		Delete: resourceNetboxPrefixDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):Allocates the next free prefix of the requested length from a parent prefix. Import with "<parent_prefix_id> <prefix_id> <prefix_length>".`,

		Schema: map[string]*schema.Schema{
			"parent_prefix_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent prefix associated with the available prefix.",
				Required:    true,
				ForceNew:    true,
			},
			"prefix_length": {
				Type:         schema.TypeInt,
				Description:  "The prefix length of the available prefix.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 128),
			},
			"prefix": {
				Type:        schema.TypeString,
				Description: "The prefix in CIDR notation.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: buildValidValueDescription(resourceNetboxPrefixStatusOptions),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the available prefix.",
				Optional:    true,
			},
			"is_pool": {
				Type:        schema.TypeBool,
				Description: "Whether all addresses of the available prefix, including network and broadcast addresses, are usable.",
				Optional:    true,
			},
			"mark_utilized": {
				Type:        schema.TypeBool,
				Description: "Whether the available prefix is treated as fully utilized.",
				Optional:    true,
			},
			"vrf_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the vrf associated with the available prefix.",
				Optional:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the available prefix.",
				Optional:    true,
			},
			"location_id": {
				Type:          schema.TypeInt,
				Description:   "The ID of the location associated with the available prefix.",
				Optional:      true,
				ConflictsWith: []string{"site_id", "site_group_id", "region_id"},
			},
			"site_id": {
				Type:          schema.TypeInt,
				Description:   "The ID of the site associated with the available prefix.",
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_group_id", "region_id"},
			},
			"site_group_id": {
				Type:          schema.TypeInt,
				Description:   "The ID of the site group associated with the available prefix.",
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_id", "region_id"},
			},
			"region_id": {
				Type:          schema.TypeInt,
				Description:   "The ID of the region associated with the available prefix.",
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_id", "site_group_id"},
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the vlan associated with the available prefix.",
				Optional:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the available prefix.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the available prefix.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the available prefix.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: func(c context.Context, rd *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parentPrefixID, prefixID, prefixLength, err := resourceNetboxAvailablePrefixParseImport(rd.Id())
				if err != nil {
					return nil, err
				}

				rd.Set("parent_prefix_id", parentPrefixID)
				rd.Set("prefix_length", prefixLength)
				rd.SetId(prefixID)

				return []*schema.ResourceData{rd}, nil
			},
		},
	}
}

func resourceNetboxAvailablePrefixParseImport(importStr string) (int, string, int, error) {
	parts := strings.SplitN(importStr, " ", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return 0, "", 0, fmt.Errorf("unexpected format of (%s), expected 'parent_prefix_id prefix_id prefix_length'", importStr)
	}

	parentID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", 0, fmt.Errorf("parent_id (%s) is not an integer", parts[0])
	}
	prefixLength, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, "", 0, fmt.Errorf("prefix_length (%s) is not an integer", parts[2])
	}

	return parentID, parts[1], prefixLength, nil
}

type availablePrefixCreateResponse struct {
	ID     int64  `json:"id"`
	Prefix string `json:"prefix"`
}

func resourceNetboxAvailablePrefixCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	parentPrefixID := int64(d.Get("parent_prefix_id").(int))

	// Send every configured field with the allocation, so NetBox rules that need
	// a tenant, role and so on on creation accept it. The prefix itself is chosen
	// by NetBox. The update below then sets whatever NetBox did not take over.
	req, err := prefixRequestFromResourceData(api, d)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	delete(data, "prefix")
	data["prefix_length"] = int64(d.Get("prefix_length").(int))

	// NetBox picks the next free prefix of the parent and returns it.
	path := fmt.Sprintf("ipam/prefixes/%d/available-prefixes/", parentPrefixID)
	payload, err := netboxapi.Request[availablePrefixCreateResponse](context.Background(), api.rest, http.MethodPost, path, nil, data)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(payload.ID, 10))
	d.Set("prefix", payload.Prefix)

	return resourceNetboxPrefixUpdate(d, m)
}
