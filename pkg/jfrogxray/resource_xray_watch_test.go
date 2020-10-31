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

const watchBasic = `
resource "xray_watch" "test" {
	name  = "test watch"
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
}`

func TestAccWatch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckUserDestroy("xray_watch.test"),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: userBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("xray_watch.test", "name", "test watch"),
				),
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

		watch, resp, err := c.V2.Watches.GetWatch(context.Backgroun(), "test watch")

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Watch %s still exists %s", rs.Primary.ID, watch)
		}
	}
}
