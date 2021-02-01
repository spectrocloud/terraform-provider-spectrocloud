resource "local_file" "import_manifest" {
  content              = local.cluster_import_manifest
  filename             = "import_manifest_aws.yaml"
  file_permission      = "0644"
  directory_permission = "0755"
}