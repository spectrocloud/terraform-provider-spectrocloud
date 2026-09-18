# Creates and manages a dedicated GCP cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained. This is the same
# cloud account type used by spectrocloud_cluster_gcp.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute updates in place.
resource "spectrocloud_cloudaccount_gcp" "account" {
  name                 = "e2e-gke-account"
  context              = "project"
  gcp_json_credentials = var.gcp_json_credentials
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_gcp" "account" {
#   name = "some-existing-account-name"
# }
