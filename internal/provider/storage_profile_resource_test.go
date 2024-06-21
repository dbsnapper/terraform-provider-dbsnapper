package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStorageProfileResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStorageProfileResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the created storage profile
					resource.TestCheckResourceAttr("dbsnapper_storage_profile.test", "name", "tf_test_storage_profile"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dbsnapper_storage_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore attributes that might not exist during import
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: testAccStorageProfileResourceConfigUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the updated storage profile
					resource.TestCheckResourceAttr("dbsnapper_storage_profile.test", "name", "tf_test_storage_profile_update"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

const testAccStorageProfileResourceConfig = `
resource "dbsnapper_storage_profile" "test" {
    name = "tf_test_storage_profile"
    sp_provider = "s3"
    region = "us-west-2"
    account_id = ""
    access_key = "AKIAxxxxxxxxxxxx"
    secret_key = "xxxxxxxxxxxxxxxxxxxx"
    bucket = "tf-test-bucket"
    prefix = "tf-test-prefix"
}
`

const testAccStorageProfileResourceConfigUpdate = `
resource "dbsnapper_storage_profile" "test" {
    name = "tf_test_storage_profile_update"
    sp_provider = "s3"
    region = "us-west-1"
    account_id = ""
    access_key = "AKIAxxxxxxxxxxxx"
    secret_key = "xxxxxxxxxxxxxxxxxxxx"
    bucket = "tf-test-bucket-updated"
    prefix = "tf-test-prefix-updated"
}
`
