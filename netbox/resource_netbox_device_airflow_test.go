package netbox

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxDevice_airflow(t *testing.T) {
	testName := testAccGetTestName("device_airflow")
	config := func(airflow string) string {
		extra := ""
		if airflow != "" {
			extra = fmt.Sprintf("airflow = %q", airflow)
		}
		return testAccNetboxDeviceFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = "%[1]s"
  device_type_id = netbox_device_type.test.id
  role_id        = netbox_device_role.test.id
  site_id        = netbox_site.test.id
  %[2]s
}`, testName, extra)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config("front-to-rear"),
				Check:  resource.TestCheckResourceAttr("netbox_device.test", "airflow", "front-to-rear"),
			},
			{
				ResourceName:      "netbox_device.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config("rear-to-front"),
				Check:  resource.TestCheckResourceAttr("netbox_device.test", "airflow", "rear-to-front"),
			},
			{
				Config: config(""),
				Check:  testAccCheckAttrUnset("netbox_device.test", "airflow"),
			},
		},
	})
}
