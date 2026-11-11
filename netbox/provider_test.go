package netbox

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"netbox": providerserver.NewProtocol6WithError(fwprovider.New("test")()),
}

var testPrefix = "test"

// testAccClient returns a client for the NetBox the acceptance tests run against.
func testAccClient() (*netboxapi.Client, error) {
	return netboxapi.New(os.Getenv("NETBOX_SERVER_URL"), os.Getenv("NETBOX_API_TOKEN"), nil)
}

func testAccGetTestName(testSlug string) string {
	randomString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	return strings.Join([]string{testPrefix, testSlug, randomString}, "-")
}

func testAccGetTestToken() string {
	randomToken := acctest.RandStringFromCharSet(40, "0123456789")
	return randomToken
}

// testAccNetboxVersionAtLeast returns true if the NETBOX_VERSION env var is >= the given constraint.
// If NETBOX_VERSION is unset or unparseable, it returns false.
func testAccNetboxVersionAtLeast(constraint string) bool {
	raw := strings.TrimPrefix(os.Getenv("NETBOX_VERSION"), "v")
	v, err := semver.NewVersion(raw)
	if err != nil {
		return false
	}
	c, err := semver.NewConstraint(">= " + constraint)
	if err != nil {
		return false
	}
	return c.Check(v)
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("NETBOX_SERVER_URL"); v == "" {
		t.Fatal("NETBOX_SERVER_URL must be set for acceptance tests.")
	}
	if v := os.Getenv("NETBOX_API_TOKEN"); v == "" {
		t.Fatal("NETBOX_API_TOKEN must be set for acceptance tests.")
	}
}

func testProviderConfig(platform string) string {
	return fmt.Sprintf(`
	resource "netbox_platform" "testplatform" {
    name = "%s"
	}`, platform)
}

func TestAccNetboxProviderConfigure_failure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "netbox" {
  server_url         = "https://fake.netbox.server"
  api_token          = "1234567890"
  skip_version_check = true
}
` + testProviderConfig(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
				ExpectError: regexp.MustCompile("(?s)fake.netbox.server.*no such host"),
			},
		},
	})
}

func TestAccNetboxProviderDefaultTags(t *testing.T) {
	defaultTag := fmt.Sprintf("managed-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "netbox" {
						default_tags = ["%[1]s"]
					}

					resource "netbox_tag" "managed" {
						name = "%[1]s"
					}

					resource "netbox_site" "testsite" {
						name = "%[2]s"

						depends_on = [
							netbox_tag.managed
						]
					}
					`, defaultTag, acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.testsite", "tags_all.#", "1"),
					resource.TestCheckResourceAttr("netbox_site.testsite", "tags_all.0", defaultTag),
				),
			},
		},
	})
}

// testAccSkipBelowVersion skips the test when the NetBox under test is older
// than the given version, for objects that NetBox added later.
func testAccSkipBelowVersion(t *testing.T, version string) {
	t.Helper()
	if !testAccNetboxVersionAtLeast(version) {
		t.Skipf("requires NetBox %s or newer", version)
	}
}

// testAccCheckIfNetboxAtLeast runs the check only when the NetBox under test is
// at least the given version, for attributes that older versions do not have.
func testAccCheckIfNetboxAtLeast(version string, check resource.TestCheckFunc) resource.TestCheckFunc {
	if !testAccNetboxVersionAtLeast(version) {
		return func(*terraform.State) error { return nil }
	}
	return check
}

// testAccCheckAttrUnset accepts both a missing attribute and a zero ID, which is
// how an unset relation is stored in the state.
func testAccCheckAttrUnset(name, attr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found", name)
		}
		if v := rs.Primary.Attributes[attr]; v != "" && v != "0" {
			return fmt.Errorf("%s.%s should be unset, got %q", name, attr, v)
		}
		return nil
	}
}

// TestProviderSchema checks that the provider server starts and reports the schemas of all
// resources and data sources without errors.
func TestProviderSchema(t *testing.T) {
	server, err := testAccProtoV6ProviderFactories["netbox"]()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		t.Errorf("%s: %s", d.Summary, d.Detail)
	}
	assert.NotEmpty(t, resp.ResourceSchemas)
}
