// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTargetsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccTargetsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbsnapper_targets.test", "targets.#", "2"),
					resource.TestCheckResourceAttr("data.dbsnapper_targets.test", "targets.0.name", "tf_target_1"),
					resource.TestCheckResourceAttr("data.dbsnapper_targets.test", "targets.0.share.sso_groups.#", "2"),
					resource.TestCheckResourceAttr("data.dbsnapper_targets.test", "targets.0.share.sso_groups.0", "target1"),
					resource.TestCheckResourceAttr("data.dbsnapper_targets.test", "targets.1.name", "tf_target_2"),
				),
			},
		},
	})
}

const testAccTargetsDataSourceConfig = `
data "dbsnapper_targets" "test" {
}
`
