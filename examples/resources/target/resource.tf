resource "dbsnapper_target" "tf_example" {
  name = "tf_example"
  snapshot = {
    src_url = "postgres://user:pass@localhost:5432/tf_example"
    dst_url = "postgres://user:pass@localhost:5432/tf_example_snap"
  }
  sanitize = {
    dst_url = "postgres://user:pass@localhost:5432/tf_example_snap_sanitized"
    query   = <<EOT
        DROP TABLE IF EXISTS dbsnapper_info;
        CREATE TABLE dbsnapper_info (created_at timestamp, tags text []);
        INSERT INTO dbsnapper_info (created_at, tags)
        VALUES (NOW(), '{target:tf-example, src:terraform}');
      EOT
  }
  share = {
    sso_groups = ["group1", "group2", "group3"]
  }
}

output "dbsnapper_target" {
  value = dbsnapper_target.tf_example
}
