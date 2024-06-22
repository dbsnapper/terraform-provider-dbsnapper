package provider

import (
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func setupSuite(t *testing.T) func(t *testing.T) {
	log.Println("setup suite")

	// Return a function to teardown the test
	return func(t *testing.T) {
		log.Println("teardown suite")
	}
}

func TestAccTargetResource(t *testing.T) {

	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTargetResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of items
					resource.TestCheckResourceAttr("dbsnapper_target.test", "name", "tf_test"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dbsnapper_target.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: testAccTargetResourceConfigUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first order item updated
					resource.TestCheckResourceAttr("dbsnapper_target.test", "name", "tf_test_update"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

const testAccTargetResourceConfig = `
resource "dbsnapper_target" "test" {
    name = "tf_test"
    snapshot = {
      src_url = "postgres://user:pass@localhost:5432/tf_test"
      dst_url = "postgres://user:pass@localhost:5432/tf_test_snap"
    }
    sanitize = {
      dst_url = "postgres://user:pass@localhost:5432/tf_test_snap_sanitized"
      query = <<EOT
        DROP TABLE IF EXISTS dbsnapper_info;
        CREATE TABLE dbsnapper_info (created_at timestamp, tags text []);
        INSERT INTO dbsnapper_info (created_at, tags)
        VALUES (NOW(), '{target:tf_test, src:terraform_test}');
      EOT
    }
    share = {
      sso_groups = ["group1", "group2", "group3"]
    }
}
`
const testAccTargetResourceConfigUpdate = `
resource "dbsnapper_target" "test" {
    name = "tf_test_update"
    snapshot = {
      src_url = "postgres://user:pass@localhost:5432/tf_test_update"
      dst_url = "postgres://user:pass@localhost:5432/tf_test_update_snap"
    }
    sanitize = {
      dst_url = "postgres://user:pass@localhost:5432/tf_test_snap_sanitized"
      query = <<EOT
        DROP TABLE IF EXISTS dbsnapper_info;
        CREATE TABLE dbsnapper_info (created_at timestamp, tags text []);
        INSERT INTO dbsnapper_info (created_at, tags)
        VALUES (NOW(), '{target:tf_test, src:terraform_test}');
      EOT
    }
    share = {
      sso_groups = ["group4", "group5", "group6"]
    }
}
`
