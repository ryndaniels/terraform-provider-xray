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
	policyName := "test policy"
	policyDesc := "policy created by xray acceptance tests"
	resourceName := "xray_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckPolicyDestroy(resourceName),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXrayPolicy_basic(policyName, policyDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", policyDesc),
					resource.TestCheckResourceAttr(resourceName, "rules.0.name", "security rule"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.criteria.min_severity", "High"),
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

func TestAccPolicy_actions(t *testing.T) {
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
}

func testAccCheckPolicyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*xray.Xray)
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		policy, resp, err := client.V1.Policies.GetPolicy(context.Background(), "test policy")

		if resp.StatusCode == http.StatusNotFound {
			return nil
		} else if err != nil {
			return fmt.Errorf("error: Request failed: %s", err.Error())
		} else {
			return fmt.Errorf("error: Policy %s still exists %s", rs.Primary.ID, *policy.Name)
		}
	}
}

func testAccXrayPolicy_basic(name, description string) string {
	return fmt.Sprintf(`
	resource "xray_policy" "test" {
		name  = "%s"
		description = "%s"
		type = "security"

		rules [
			name = "security rule"
			priority = 1
			criteria [
				min_severity = "High"
			]
		]
	}
`, name, description)
}

func testAccXrayPolicy_actions(name, description, policyType, failBuild string) string {
	return fmt.Sprintf(`
	resource "xray_policy" "test" {
		name  = "%s"
		description = "%s"
		type = "%s"
		author = "Acctest Author"

		rules [
			{
				name = "security rule"
				priority = 1
				criteria [
					min_severity = "High"
				]
			}
		]

		actions [
			{
				fail_build = "%s"
			}
		]
		
	}
`, name, description, policyType, failBuild)
}
