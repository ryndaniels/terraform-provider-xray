package jfrogxray

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"xray": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("XRAY_URL"); v == "" {
		t.Fatal("XRAY_URL must be set for acceptance tests")
	}

	username := os.Getenv("XRAY_USERNAME")
	password := os.Getenv("XRAY_PASSWORD")
	accessToken := os.Getenv("XRAY_ACCESS_TOKEN")

	if (username == "" || password == "") && accessToken == "" {
		t.Fatal("either XRAY_USERNAME/XRAY_PASSWORD or XRAY_ACCESS_TOKEN must be set for acceptance test")
	}

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}
