cluster_cloud_account_name   = "gcp-1"
cluster_cluster_profile_name = "ProdGoogle"

gcp_serviceaccount_json = <<-EOT
  {
    "type": "service_account",
    "project_id": "gcp-project-1",
    ...
  }
EOT
