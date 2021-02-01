resource "local_file" "import_manifest1" {
  content              = local.cluster_import_manifest
  filename             = "import_manifest_vsphere.yaml"
  file_permission      = "0644"
  directory_permission = "0755"
}