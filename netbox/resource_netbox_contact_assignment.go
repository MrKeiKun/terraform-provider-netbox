package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxContactAssignmentPriorityOptions = []string{"primary", "secondary", "tertiary", "inactive"}

func resourceNetboxContactAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxContactAssignmentCreate,
		Read:   resourceNetboxContactAssignmentRead,
		Update: resourceNetboxContactAssignmentUpdate,
		Delete: resourceNetboxContactAssignmentDelete,

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/contacts#contactassignments_1):

> Much like tenancy, contact assignment enables you to track ownership of resources modeled in NetBox.`,

		Schema: map[string]*schema.Schema{
			"content_type": {
				Type:        schema.TypeString,
				Description: "The content type of the contact assignment.",
				Required:    true,
			},
			"object_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the object associated with the contact assignment.",
				Required:    true,
			},
			"contact_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the contact associated with the contact assignment.",
				Required:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the contact assignment.",
				Required:    true,
			},
			"priority": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxContactAssignmentPriorityOptions, false),
				Description:  buildValidValueDescription(resourceNetboxContactAssignmentPriorityOptions),
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const contactAssignmentAPIPath = "tenancy/contact-assignments/"

type contactAssignmentRequest struct {
	ObjectType   string      `json:"object_type,omitempty"`
	ObjectId     *int64      `json:"object_id"`
	Contact      *int64      `json:"contact"`
	Role         *int64      `json:"role"`
	Priority     string      `json:"priority,omitempty"`
	CustomFields interface{} `json:"custom_fields,omitempty"`
}

type contactAssignmentResponse struct {
	ObjectType   string                 `json:"object_type"`
	ObjectId     *int64                 `json:"object_id"`
	Contact      *netboxapi.Ref         `json:"contact"`
	Role         *netboxapi.Ref         `json:"role"`
	Priority     *netboxapi.Choice      `json:"priority"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func contactAssignmentRequestFromResourceData(d *schema.ResourceData) *contactAssignmentRequest {
	data := &contactAssignmentRequest{
		CustomFields: customFieldsRequest(d),
		ObjectType:   d.Get("content_type").(string),
		ObjectId:     int64ToPtr(int64(d.Get("object_id").(int))),
		Contact:      int64ToPtr(int64(d.Get("contact_id").(int))),
		Role:         int64ToPtr(int64(d.Get("role_id").(int))),
		Priority:     d.Get("priority").(string),
	}
	return data
}

func resourceNetboxContactAssignmentCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, contactAssignmentAPIPath, contactAssignmentRequestFromResourceData(d), resourceNetboxContactAssignmentRead)
}

func resourceNetboxContactAssignmentRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, contactAssignmentAPIPath, func(r *contactAssignmentResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("content_type", r.ObjectType)
		d.Set("object_id", r.ObjectId)
		if r.Contact != nil {
			d.Set("contact_id", r.Contact.ID)
		}
		if r.Role != nil {
			d.Set("role_id", r.Role.ID)
		}
		d.Set("priority", netboxapi.ChoiceValue(r.Priority))
	})
}

func resourceNetboxContactAssignmentUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, contactAssignmentAPIPath, contactAssignmentRequestFromResourceData(d), resourceNetboxContactAssignmentRead)
}

func resourceNetboxContactAssignmentDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, contactAssignmentAPIPath)
}
