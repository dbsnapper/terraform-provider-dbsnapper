terraform {
  required_providers {
    dbsnapper = {
      source = "registry.terraform.io/hashicorp/dbsnapper"
    }
  }
}

provider "dbsnapper" {}

data "dbsnapper_targets" "example" {}
