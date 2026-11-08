package netbox

import (
	"context"
	"regexp"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxTagCreate,
		Read:   resourceNetboxTagRead,
		Update: resourceNetboxTagUpdate,
		Delete: resourceNetboxTagDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/extras/tag/):
> Tags are user-defined labels which can be applied to a variety of objects within NetBox. They can be used to establish dimensions of organization beyond the relationships built into NetBox. For example, you might create a tag to identify a particular ownership or condition across several types of objects.
>
> Each tag has a label, color, and a URL-friendly slug. For example, the slug for a tag named "Dunder Mifflin, Inc." would be dunder-mifflin-inc. The slug is generated automatically and makes tags easier to work with as URL parameters. Each tag can also be assigned a description indicating its purpose.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the tag.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the tag.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"color_hex": {
				Type:         schema.TypeString,
				Description:  "The color hex of the tag.",
				Optional:     true,
				Default:      "9e9e9e",
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-f]{6}$"), "Must be hex color string"),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the tag.",
				Optional:    true,
			},
			"weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
				Description:  "Weight of the tag. Tags with a higher weight are shown after tags with a lower weight.",
			},
			"object_types": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Restricts the tag to these object types, for example `dcim.site`. The tag can be used on every object type when this is empty.",
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// tagAPIPath is the NetBox API collection path for tags.
const tagAPIPath = "extras/tags/"

// netboxTag is the subset of the NetBox tag object this resource uses.
type netboxTag struct {
	ID          int64    `json:"id,omitempty"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Color       string   `json:"color"`
	Description string   `json:"description"`
	Weight      *int64   `json:"weight,omitempty"`
	ObjectTypes []string `json:"object_types"`
}

func tagFromResourceData(d *schema.ResourceData) *netboxTag {
	name := d.Get("name").(string)

	// Default slug to generated slug if not given
	slug := getSlug(name)
	if v, ok := d.GetOk("slug"); ok {
		slug = v.(string)
	}

	return &netboxTag{
		Name:        name,
		Slug:        slug,
		Color:       d.Get("color_hex").(string),
		Description: d.Get("description").(string),
		Weight:      getOptionalInt(d, "weight"),
		ObjectTypes: getStringSet(d, "object_types"),
	}
}

func resourceNetboxTagCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	res, err := netboxapi.Create[netboxTag](context.Background(), api.rest, tagAPIPath, tagFromResourceData(d))
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.ID, 10))

	return resourceNetboxTagRead(d, m)
}

func resourceNetboxTagRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	res, err := netboxapi.Get[netboxTag](context.Background(), api.rest, tagAPIPath, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.Name)
	d.Set("slug", res.Slug)
	d.Set("color_hex", res.Color)
	d.Set("description", res.Description)
	d.Set("weight", res.Weight)
	d.Set("object_types", res.ObjectTypes)
	return nil
}

func resourceNetboxTagUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if _, err := netboxapi.Update[netboxTag](context.Background(), api.rest, tagAPIPath, id, tagFromResourceData(d)); err != nil {
		return err
	}

	return resourceNetboxTagRead(d, m)
}

func resourceNetboxTagDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if err := netboxapi.Delete(context.Background(), api.rest, tagAPIPath, id); err != nil {
		if netboxapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	return nil
}
