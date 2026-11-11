package netbox

import (
	"fmt"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetboxClusterGroup_basic(t *testing.T) {
	testSlug := "clstrgrp_basic"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_cluster_group" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  description = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cluster_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cluster_group.test", "slug", testName),
					resource.TestCheckResourceAttr("netbox_cluster_group.test", "description", testName),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "netbox_cluster_group" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  description = "%[1]s-updated"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_cluster_group.test", "description", regexp.MustCompile(`-updated$`)),
				),
			},
			{
				ResourceName:      "netbox_cluster_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetboxClusterGroup_defaultSlug(t *testing.T) {
	testSlug := "clstrgrp_defSlug"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_cluster_group" "test" {
  name = "%[1]s"
  description = "%[1]s"
}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_cluster_group.test", "name", testName),
					resource.TestCheckResourceAttr("netbox_cluster_group.test", "slug", fwresource.Slug(testName)),
					resource.TestCheckResourceAttr("netbox_cluster_group.test", "description", testName),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_cluster_group", &resource.Sweeper{
		Name:         "netbox_cluster_group",
		Dependencies: []string{},
		F: func(region string) error {
			return sweepByName(region, "virtualization/cluster-groups/", "name")
		},
	})
}
