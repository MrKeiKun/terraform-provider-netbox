package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccNetboxVirtualMachineTypeDependencies(testName string) string {
	_ = testName
	return ""
}

func TestAccNetboxVirtualMachineType_basic(t *testing.T) {
	testName := testAccGetTestName("virtualmachine")
	deps := testAccNetboxVirtualMachineTypeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.6.0")
		},
		CheckDestroy: func(s *terraform.State) error {
			return testAccCheckObjectsDestroyed(s, "netbox_virtual_machine_type", "virtualization/virtual-machine-types/")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_machine_type" "test" {
  name = "%[1]s"
  default_vcpus = 2
  default_memory = 2048

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_type.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_virtual_machine_type.test", "default_vcpus", "2"),
					resource.TestCheckResourceAttr("netbox_virtual_machine_type.test", "default_memory", "2048"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine_type.test", "slug"),
				),
			},
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_machine_type" "test" {
  description = "updated"
  name = "%[1]s"
  default_vcpus = 2
  default_memory = 2048

}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_type.test", "description", "updated"),
				),
			},
			{
				ResourceName:      "netbox_virtual_machine_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxVirtualMachineTypeDataSource_basic(t *testing.T) {
	testName := testAccGetTestName("virtualmachine_ds")
	deps := testAccNetboxVirtualMachineTypeDependencies(testName)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			testAccSkipBelowVersion(t, "4.6.0")
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`%[2]s
resource "netbox_virtual_machine_type" "test" {
  name = "%[1]s"
  default_vcpus = 2
  default_memory = 2048

}

data "netbox_virtual_machine_type" "test" {
  depends_on = [netbox_virtual_machine_type.test]
  name = "%[1]s"
}`, testName, deps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_virtual_machine_type.test", "id", "netbox_virtual_machine_type.test", "id"),
					resource.TestCheckResourceAttrPair("data.netbox_virtual_machine_type.test", "name", "netbox_virtual_machine_type.test", "name"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_virtual_machine_type", &resource.Sweeper{
		Name:         "netbox_virtual_machine_type",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "virtualization/virtual-machine-types/", "name")
		},
	})
}
