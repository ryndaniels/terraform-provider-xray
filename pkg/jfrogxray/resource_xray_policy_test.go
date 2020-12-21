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

func TestAccPolicy_basic(t *testing.T) {
	policyName   := "terraform-test-policy"
	policyDesc   := "policy created by xray acceptance tests"
	ruleName     := "test-security-rule"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy,
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_basic(policyName, policyDesc, ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					//resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.min_severity", "High"),
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

/*func TestAccPolicy_actions(t *testing.T) {
	policyName := "test policy"
	policyDesc := "policy created by xray acceptance tests"
	policyType := "security"
	failBuild := "true"
	updatedDesc := "updated policy description"
	updatedType := "license"
	updatedFailBuild := "false"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy(resourceName),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_actions(policyName, policyDesc, policyType, failBuild),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", "security rule"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.min_severity", "High"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.min_severity", "High"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.fail_build", failBuild),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccXrayPolicy_actions(policyName, updatedDesc, updatedType, updatedFailBuild),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", "security rule"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.min_severity", "High"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.min_severity", "High"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.fail_build", updatedFailBuild),
				),
			},
		},
	})
}*/

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*xray.Xray)

	for _, rs := range s.RootModule().Resources {
		fmt.Printf("resource type is %v\n", rs.Type)
		if rs.Type != "xray_policy" {
			continue
		}

		policy, resp, err := conn.V1.Policies.GetPolicy(context.Background(), rs.Primary.ID)

		if resp.StatusCode == http.StatusNotFound {
			continue
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Policy %s still exists %s", rs.Primary.ID, *policy.Name)
		}
	}
	return nil
}

func testAccXrayPolicy_basic(name, description, ruleName string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "%s"
	type = "security"

	rules {
		name = "%s"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			fail_build = true
		}
	}
}
`, name, description, ruleName)
}

func testAccXrayPolicy_actions(name, description, policyType, failBuild string) string {
	return fmt.Sprintf(`
resource "xray_policy" "test" {
	name  = "%s"
	description = "%s"
	type = "%s"
	author = "Acctest Author"

	rules {
		name = "security rule"
		priority = 1
		criteria {
			min_severity = "High"
		}
		actions {
			fail_build = "%s"
		}
	}
}
`, name, description, policyType, failBuild)
}

// TODO write a _full test with all criteria, all actions