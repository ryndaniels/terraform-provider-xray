package jfrogxray

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xero-oss/go-xray/xray"
)

func TestAccWatch_basic(t *testing.T) {
	watchName := "test watch"
	watchDesc := "watch created by xray acceptance tests"
	resourceName := "xray_watch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy(resourceName),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayWatch_Basic(watchName, watchDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", watchDesc),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckWatchDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*xray.Xray)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		watch, resp, err := client.V2.Watches.GetWatch(context.Background(), "test watch")

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Watch %s still exists %s", rs.Primary.ID, *watch.GeneralData.Name)
		}
	}
}

func testAccXrayWatch_Basic(name, description string) string {
	return fmt.Sprintf(`
	resource "xray_watch" "test" {
		name  = "%s"
		description = "%s"
		resources [
			{
				type = "repository"
				bin_mgr_id = "123456"
			}
		]
		assigned_policies [
			{
				name = "policy_name"
				type = "security
			}
		]
	}
`, name, description)
}
