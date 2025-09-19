package netbox

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client/circuits"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxCircuitTermination_basic(t *testing.T) {
	testSlug := "circuit_term"
	testName := testAccGetTestName(testSlug)
	randomSlug := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[2]s"
  status = "active"
}
resource "netbox_circuit_provider" "test" {
  name = "%[1]s"
  slug = "%[2]s"
}
resource "netbox_circuit_type" "test" {
  name = "%[1]s"
  slug = "%[2]s"
}
resource "netbox_circuit" "test" {
  cid = "%[1]s"
  status = "active"
  provider_id = netbox_circuit_provider.test.id
  type_id = netbox_circuit_type.test.id
}
resource "netbox_circuit_termination" "test" {
  circuit_id = netbox_circuit.test.id
  term_side = "A"
  site_id = netbox_site.test.id
  port_speed = 100000
  upstream_speed = 50000
}`, testName, randomSlug),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_circuit_termination.test", "circuit_id", "netbox_circuit.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_circuit_termination.test", "site_id", "netbox_site.test", "id"),
					resource.TestCheckResourceAttr("netbox_circuit_termination.test", "port_speed", "100000"),
					resource.TestCheckResourceAttr("netbox_circuit_termination.test", "upstream_speed", "50000"),
				),
			},
			{
				ResourceName:      "netbox_circuit_termination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxCircuitTermination_providerNetwork(t *testing.T) {
	testSlug := "circuit_term_pn"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_circuit_provider" "test_provider" {
  name = "%[1]s_provider"
}
resource "netbox_provider_network" "test_network" {
  name        = "%[1]s_network"
  provider_id = netbox_circuit_provider.test_provider.id
  service_id  = "SVC001"
}
resource "netbox_circuit_type" "test" {
  name = "%[1]s_type"
}
resource "netbox_circuit" "test" {
  cid         = "%[1]s_circuit"
  status      = "active"
  provider_id = netbox_circuit_provider.test_provider.id
  type_id     = netbox_circuit_type.test.id
}
resource "netbox_circuit_termination" "test" {
  circuit_id          = netbox_circuit.test.id
  term_side           = "A"
  provider_network_id = netbox_provider_network.test_network.id
  port_speed          = 100000
  upstream_speed      = 50000
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("netbox_circuit_termination.test", "circuit_id", "netbox_circuit.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_circuit_termination.test", "provider_network_id", "netbox_provider_network.test_network", "id"),
					resource.TestCheckResourceAttr("netbox_circuit_termination.test", "port_speed", "100000"),
					resource.TestCheckResourceAttr("netbox_circuit_termination.test", "upstream_speed", "50000"),
				),
			},
			{
				ResourceName:      "netbox_circuit_termination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_circuit_termination", &resource.Sweeper{
		Name:         "netbox_circuit_termination",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*providerState)
			params := circuits.NewCircuitsCircuitsListParams()
			res, err := api.Circuits.CircuitsCircuitsList(params, nil)
			if err != nil {
				return err
			}
			for _, Circuit := range res.GetPayload().Results {
				if strings.HasPrefix(*Circuit.Cid, testPrefix) {
					deleteParams := circuits.NewCircuitsCircuitsDeleteParams().WithID(Circuit.ID)
					_, err := api.Circuits.CircuitsCircuitsDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted a circuit termination")
				}
			}
			return nil
		},
	})
}
