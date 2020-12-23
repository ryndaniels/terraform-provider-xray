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
	watchName := "test-watch"
	policyName := "test-policy"
	watchDesc := "watch created by xray acceptance tests"
	resourceName := "xray_watch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayWatch_basic(watchName, watchDesc, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", watchDesc),
					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "all-repos"),
					resource.TestCheckResourceAttr(resourceName, "assigned_policies.0.name", policyName),
					resource.TestCheckResourceAttr(resourceName, "assigned_policies.0.type", "security"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayWatch_unassigned(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWatchDoesntExist(resourceName),
				),
			},
		},
	})
}

// TODO can't do filters with all-repos apparently - it's an example in the docs but impossible via the web ui
/*func TestAccWatch_filters(t *testing.T) {
	watchName := "test-watch"
	watchDesc := "watch created by xray acceptance tests"
	policyName := "test-policy"
	filterValue := "Debian"
	updatedDesc := "updated watch description"
	updatedValue := "Docker"
	resourceName := "xray_watch.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckWatchDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayWatch_filters(watchName, watchDesc, policyName, filterValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", watchDesc),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.type", "package-type"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.value", filterValue),
					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "repository"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccXrayWatch_filters(watchName, updatedDesc, policyName, updatedValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", watchName),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDesc),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.type", "package-type"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.filters.0.value", updatedValue),
					resource.TestCheckResourceAttr(resourceName, "resources.0.type", "repository"),
				),
			},
			{
				Config: testAccXrayWatch_unassigned(policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWatchDoesntExist(resourceName),
				),
			},
		},
	})
}*/

func testAccCheckWatchDoesntExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if ok {
			return fmt.Errorf("Watch %s exists when it shouldn't", resourceName)
		}
		return nil
	}
}

func testAccCheckWatchDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*xray.Xray)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "xray_watch" {
			watch, resp, err := conn.V2.Watches.GetWatch(context.Background(), rs.Primary.ID)
			if resp.StatusCode == http.StatusNotFound {
				continue
			} else if err != nil {
				return fmt.Errorf("error: Request failed: %s", err.Error())
			} else {
				return fmt.Errorf("error: Watch %s still exists %s", rs.Primary.ID, *watch.GeneralData.Name)
			}
		} else if rs.Type == "xray_policy" {
			policy, resp, err := conn.V1.Policies.GetPolicy(context.Background(), rs.Primary.ID)

			if resp.StatusCode == http.StatusNotFound {
				continue
			} else if resp.StatusCode == http.StatusInternalServerError && err.Error() == fmt.Sprintf("{\"error\":\"Failed to find Policy %s\"}", rs.Primary.ID) {
				continue
			} else if err != nil {
				return fmt.Errorf("error: Request failed: %s", err.Error())
			} else {
				return fmt.Errorf("error: Policy %s still exists %s", rs.Primary.ID, *policy.Name)
			}
		} else {
			continue
		}
		
	}

	return nil
}

func testAccXrayWatch_basic(name, description, policyName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "test policy description"
	type = "security"

	rules {
		name = "rule-name"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "xray_watch" "test" {
	name  = "%s"
	description = "%s"
	resources {
		type = "all-repos"
		name = "All Repositories"
	}
	assigned_policies {
		name = xray_policy.test.name
		type = "security"
	}
}
`, policyName, name, description)
}

// Since policies can't be deleted if they have a watch assigned, we need to force terraform to delete the watch first
// by removing it from the code at the end of every test step
func testAccXrayWatch_unassigned(policyName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "test policy description"
	type = "security"

	rules {
		name = "rule-name"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}
`, policyName)
}

func testAccXrayWatch_filters(name, description, policyName, filterValue string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "test policy description"
	type = "security"

	rules {
		name = "rule-name"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			block_download {
				unscanned = true
				active = true
			}
		}
	}
}

resource "xray_watch" "test" {
	name  = "%s"
	description = "%s"
	resources {
		type = "all-repos"
		name = "All Repositories"
		filters {
			type = "package-type"
			value = "%s"
		}
	}
	assigned_policies {
		name = xray_policy.test.name
		type = "security"
	}
}
`,policyName, name, description, filterValue)
}

// TODO single-repo watch - won't be runnable publicly since depends on actual repos
// TODO watch specific build - probably also won't be testable publicly
// TODO test watch_recipients
