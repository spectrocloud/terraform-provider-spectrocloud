# Imports a cluster profile definition (exported from Palette: Profile > ... > Export) as a new
# cluster profile. Nothing on this resource is ForceNew, but note the provider only accepts a
# path that resolves within Terraform's own working directory (no absolute paths outside it, no
# "..") - keep the export file alongside your .tf files, as shown below.
resource "spectrocloud_cluster_profile_import" "import" {
  # Required. Path to the exported cluster profile file, resolved relative to the directory
  # `terraform apply` is run from.
  import_file = "./profile_import.json"

  # Optional, default "project". Allowed: "project", "tenant", "system".
  context = "project"
}

# Note: this resource does not support `terraform import` - it only creates a new cluster
# profile from a local export file. To manage an existing cluster profile with Terraform
# instead, use the spectrocloud_cluster_profile resource (which does support import).
