package netbox

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client/circuits"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxProviderNetwork_basic(t *testing.T) {
	testSlug := "prov_net"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_provider" "test_provider" {
  name = "%[1]s"
}

resource "netbox_provider_network" "test" {
  name         = "%[2]s"
  provider_id  = netbox_circuit_provider.test_provider.id
  service_id   = "SVC001"
  description  = "Test provider network"
}`, testName+"_provider", testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_provider_network.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_provider_network.test", "service_id", "SVC001"),
					resource.TestCheckResourceAttr("netbox_provider_network.test", "description", "Test provider network"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_provider" "test_provider" {
  name = "%[1]s"
}

resource "netbox_provider_network" "test" {
  name         = "%[2]s"
  provider_id  = netbox_circuit_provider.test_provider.id
  service_id   = "SVC002"
  description  = "Updated test provider network"
}`, testName+"_provider", testName+"2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_provider_network.test", "name", testName+"2"),
					resource.TestCheckResourceAttr("netbox_provider_network.test", "service_id", "SVC002"),
					resource.TestCheckResourceAttr("netbox_provider_network.test", "description", "Updated test provider network"),
				),
			},
			{
				ResourceName:      "netbox_provider_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_provider_network", &resource.Sweeper{
		Name:         "netbox_provider_network",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*providerState)
			params := circuits.NewCircuitsProviderNetworksListParams()
			res, err := api.Circuits.CircuitsProviderNetworksList(params, nil)
			if err != nil {
				return err
			}
			for _, ProviderNetwork := range res.GetPayload().Results {
				if strings.HasPrefix(*ProviderNetwork.Name, testPrefix) {
					deleteParams := circuits.NewCircuitsProviderNetworksDeleteParams().WithID(ProviderNetwork.ID)
					_, err := api.Circuits.CircuitsProviderNetworksDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted a provider network")
				}
			}
			return nil
		},
	})
}
