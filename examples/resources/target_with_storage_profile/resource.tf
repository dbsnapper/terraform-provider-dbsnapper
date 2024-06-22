resource "dbsnapper_storage_profile" "tfex" {
  name = "tf_sp_tfex"
  sp_provider = "s3"
  region = "us-east-1"
  account_id = ""
  access_key = "AKIAxxxxxxxx"
  secret_key = "xxxxxxxxxxxxxxxxxxxx"
  bucket = "dbsnapper-test-s3"
  prefix = "terraform"
}

resource "dbsnapper_target" "tfex" {
  name = "tf_target_sp_tfex"
  snapshot = {
    src_url = "postgres://user:pass@localhost:5432/tf_example"
    dst_url = "postgres://user:pass@localhost:5432/tf_example_snap"
    storage_profile = {
      id = dbsnapper_storage_profile.tfex.id
    }
  }
  sanitize = {
    dst_url = "postgres://user:pass@localhost:5432/tf_example_snap_sanitized"
    query   = <<EOT
        DROP TABLE IF EXISTS dbsnapper_info;
        CREATE TABLE dbsnapper_info (created_at timestamp, tags text []);
        INSERT INTO dbsnapper_info (created_at, tags)
        VALUES (NOW(), '{target:tf-example, src:terraform}');
      EOT
    storage_profile = {
      id = dbsnapper_storage_profile.tfex.id
    }
  }
  share = {
    sso_groups = ["group1", "group2", "group3"]
  }
}

output "target" {
  value = dbsnapper_target.tfex
}

output "sp_id" {
  value = dbsnapper_storage_profile.tfex.id
}
