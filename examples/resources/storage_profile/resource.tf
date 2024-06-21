resource "dbsnapper_storage_profile" "tf_sp_example" {
  name = "tf_sp_example"
  sp_provider = "s3" # s3, r2

  region = "us-east-1"
  account_id = "" # for cloudflare
  
  access_key = "AKIAxxxxxxxxxxxx"
  secret_key = "xxxxxxxxxxxxxxxxxxxx"

  bucket = "dbsnapper-test-s3"
  prefix = "terraform"
}

output "sp_id" {
  value = dbsnapper_storage_profile.tf_sp_example.id
}
output "sp_status" {
  value = dbsnapper_storage_profile.tf_sp_example.status
}
