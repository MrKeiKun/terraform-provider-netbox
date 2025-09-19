package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client/circuits"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxProviderNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxProviderNetworkCreate,
		Read:   resourceNetboxProviderNetworkRead,
		Update: resourceNetboxProviderNetworkUpdate,
		Delete: resourceNetboxProviderNetworkDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#provider-networks):

> A provider network is a network that is provided by a circuit provider. Provider networks can be used to represent networks that are available through a provider's circuits, such as MPLS VPNs or Internet transit services.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxProviderNetworkCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	data := models.WritableProviderNetwork{}

	name := d.Get("name").(string)
	data.Name = &name

	providerID := d.Get("provider_id").(int)
	data.Provider = int64ToPtr(int64(providerID))

	serviceIDValue, ok := d.GetOk("service_id")
	if ok {
		data.ServiceID = serviceIDValue.(string)
	}

	descriptionValue, ok := d.GetOk("description")
	if ok {
		data.Description = descriptionValue.(string)
	}

	data.Tags = []*models.NestedTag{}
	data.Comments = ""

	var err error
	data.Tags, err = getNestedTagListFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return err
	}

	cf, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = cf
	}

	params := circuits.NewCircuitsProviderNetworksCreateParams().WithData(&data)

	res, err := api.Circuits.CircuitsProviderNetworksCreate(params, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxProviderNetworkRead(d, m)
}

func resourceNetboxProviderNetworkRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := circuits.NewCircuitsProviderNetworksReadParams().WithID(id)

	res, err := api.Circuits.CircuitsProviderNetworksRead(params, nil)

	if err != nil {
		if errresp, ok := err.(*circuits.CircuitsProviderNetworksReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("name", res.GetPayload().Name)
	if res.GetPayload().Provider != nil {
		d.Set("provider_id", res.GetPayload().Provider.ID)
	}
	d.Set("service_id", res.GetPayload().ServiceID)
	d.Set("description", res.GetPayload().Description)

	api.readTags(d, res.GetPayload().Tags)

	cf := getCustomFields(res.GetPayload().CustomFields)
	if cf != nil {
		d.Set(customFieldsKey, cf)
	}

	return nil
}

func resourceNetboxProviderNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.WritableProviderNetwork{}

	name := d.Get("name").(string)
	data.Name = &name

	providerID := d.Get("provider_id").(int)
	data.Provider = int64ToPtr(int64(providerID))

	serviceIDValue, ok := d.GetOk("service_id")
	if ok {
		data.ServiceID = serviceIDValue.(string)
	}

	descriptionValue, ok := d.GetOk("description")
	if ok {
		data.Description = descriptionValue.(string)
	}

	data.Tags = []*models.NestedTag{}
	data.Comments = ""

	var err error
	data.Tags, err = getNestedTagListFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return err
	}

	cf, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = cf
	}

	params := circuits.NewCircuitsProviderNetworksPartialUpdateParams().WithID(id).WithData(&data)

	_, err = api.Circuits.CircuitsProviderNetworksPartialUpdate(params, nil)
	if err != nil {
		return err
	}

	return resourceNetboxProviderNetworkRead(d, m)
}

func resourceNetboxProviderNetworkDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := circuits.NewCircuitsProviderNetworksDeleteParams().WithID(id)

	_, err := api.Circuits.CircuitsProviderNetworksDelete(params, nil)
	if err != nil {
		if errresp, ok := err.(*circuits.CircuitsProviderNetworksDeleteDefault); ok {
			if errresp.Code() == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	return nil
}
