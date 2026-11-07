package netbox

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// sweeperNetboxClients is a shared cache of netbox clients
// This prevents client re-initialization for every resource with no benefit
var sweeperNetboxClients map[string]interface{}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedClientForRegion returns a common provider state configured for the specified region
func sharedClientForRegion(region string) (interface{}, error) {
	if client, ok := sweeperNetboxClients[region]; ok {
		return client, nil
	}

	c, err := netboxapi.New(os.Getenv("NETBOX_SERVER_URL"), os.Getenv("NETBOX_API_TOKEN"), nil)
	if err != nil {
		return nil, err
	}

	return &providerState{rest: c}, nil
}

// sweepByName deletes every object under path whose nameField starts with testPrefix.
func sweepByName(region, path, nameField string) error {
	m, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	api := m.(*providerState)

	objects, err := netboxapi.List[map[string]interface{}](context.Background(), api.rest, path, nil)
	if err != nil {
		return err
	}
	for _, obj := range objects {
		name, _ := obj[nameField].(string)
		if !strings.HasPrefix(name, testPrefix) {
			continue
		}
		id, ok := obj["id"].(float64)
		if !ok {
			continue
		}
		if err := netboxapi.Delete(context.Background(), api.rest, path, int64(id)); err != nil {
			return err
		}
		log.Printf("[DEBUG] Deleted %s %q", path, name)
	}
	return nil
}

// testAccCheckObjectsDestroyed is the destroy check of the acceptance tests: no
// object of resourceType in the state may still exist in NetBox under path.
func testAccCheckObjectsDestroyed(s *terraform.State, resourceType, path string) error {
	api := testAccProvider.Meta().(*providerState)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		id, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)
		_, err := netboxapi.Get[netboxapi.Ref](context.Background(), api.rest, path, id)
		if err == nil {
			return fmt.Errorf("%s (%s) still exists", resourceType, rs.Primary.ID)
		}
		if !netboxapi.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// sweepAll deletes every object under path, for objects that have no name to match on.
func sweepAll(region, path string) error {
	m, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	api := m.(*providerState)

	objects, err := netboxapi.List[map[string]interface{}](context.Background(), api.rest, path, nil)
	if err != nil {
		return err
	}
	for _, obj := range objects {
		id, ok := obj["id"].(float64)
		if !ok {
			continue
		}
		if err := netboxapi.Delete(context.Background(), api.rest, path, int64(id)); err != nil {
			return err
		}
		log.Printf("[DEBUG] Deleted %s %d", path, int64(id))
	}
	return nil
}
