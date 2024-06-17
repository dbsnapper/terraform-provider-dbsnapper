
data "dbsnapper_targets" "example" {}


output "example_targets" {
  value = data.dbsnapper_targets.example
}


