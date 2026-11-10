package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxContact() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxContactCreate),
		ReadContext:   withContext(resourceNetboxContactRead),
		UpdateContext: withContext(resourceNetboxContactUpdate),
		DeleteContext: withContext(resourceNetboxContactDelete),

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/contacts/#contacts_1):

> A contact should represent an individual or permanent point of contact. Each contact must define a name, and may optionally include a title, phone number, email address, and related details.
>
> Contacts are reused for assignments, so each unique contact must be created only once and can be assigned to any number of NetBox objects, and there is no limit to the number of assigned contacts an object may have. Most core objects in NetBox can have contacts assigned to them.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the contact.",
				Required:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the contact.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the contact.",
				Optional:    true,
			},
			"title": {
				Type:        schema.TypeString,
				Description: "The title of the contact.",
				Optional:    true,
			},
			"address": {
				Type:        schema.TypeString,
				Description: "The address of the contact.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the contact.",
				Optional:    true,
				Deprecated:  "This field is deprecated. Please use the new \"group_ids\" attribute instead.",
			},
			"group_ids": {
				Type:        schema.TypeList,
				Description: "The IDs of the group associated with the contact.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional: true,
			},
			"email": {
				Type:        schema.TypeString,
				Description: "The email address of the contact.",
				Optional:    true,
			},
			"phone": {
				Type:        schema.TypeString,
				Description: "The phone number of the contact.",
				Optional:    true,
			},
			"link": {
				Type:        schema.TypeString,
				Description: "A link associated with the contact.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the contact.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const contactAPIPath = "tenancy/contacts/"

type contactRequest struct {
	Name         string             `json:"name"`
	Phone        string             `json:"phone,omitempty"`
	Email        string             `json:"email,omitempty"`
	Link         string             `json:"link,omitempty"`
	Description  string             `json:"description,omitempty"`
	Groups       []int64            `json:"groups"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Title        string             `json:"title,omitempty"`
	Address      string             `json:"address,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type contactResponse struct {
	Name         string                 `json:"name"`
	Phone        string                 `json:"phone"`
	Email        string                 `json:"email"`
	Link         string                 `json:"link"`
	Description  string                 `json:"description"`
	Groups       []netboxapi.Ref        `json:"groups"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Title        string                 `json:"title"`
	Address      string                 `json:"address"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func contactRequestFromResourceData(api *providerState, d *schema.ResourceData) (*contactRequest, error) {
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))

	groups := make([]int64, 0)
	if groupIDs := d.Get("group_ids").([]interface{}); len(groupIDs) > 0 {
		for _, v := range groupIDs {
			groups = append(groups, int64(v.(int)))
		}
	} else if groupID := int64(d.Get("group_id").(int)); groupID != 0 {
		groups = append(groups, groupID)
	}

	return &contactRequest{
		CustomFields: customFieldsRequest(d),
		Address:      getOptionalStr(d, "address", true),
		Title:        getOptionalStr(d, "title", true),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         d.Get("name").(string),
		Phone:        d.Get("phone").(string),
		Email:        d.Get("email").(string),
		Link:         d.Get("link").(string),
		Description:  d.Get("description").(string),
		Groups:       groups,
		Tags:         tags,
	}, nil
}

func resourceNetboxContactCreate(d *schema.ResourceData, m interface{}) error {
	data, err := contactRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, contactAPIPath, data, resourceNetboxContactRead)
}

func resourceNetboxContactRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, contactAPIPath, func(contact *contactResponse) {
		d.Set(customFieldsKey, getCustomFields(contact.CustomFields))
		d.Set("address", contact.Address)
		d.Set("title", contact.Title)
		d.Set("comments", contact.Comments)
		d.Set("owner_id", netboxapi.RefID(contact.Owner))
		d.Set("name", contact.Name)
		d.Set("phone", contact.Phone)
		d.Set("email", contact.Email)
		d.Set("link", contact.Link)
		d.Set("description", contact.Description)

		if contact.Groups != nil {
			groupIDs := netboxapi.IDs(contact.Groups)
			if _, ok := d.GetOk("group_ids"); ok {
				d.Set("group_ids", groupIDs)
				d.Set("group_id", nil)
			} else if _, ok := d.GetOk("group_id"); ok && len(groupIDs) == 1 {
				d.Set("group_id", groupIDs[0])
				d.Set("group_ids", nil)
			}
		} else {
			d.Set("group_ids", nil)
			d.Set("group_id", nil)
		}
	})
}

func resourceNetboxContactUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := contactRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, contactAPIPath, data, resourceNetboxContactRead)
}

func resourceNetboxContactDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, contactAPIPath)
}
