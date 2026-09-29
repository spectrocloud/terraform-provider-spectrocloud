# Creates and manages a dedicated MAAS cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained and every
# attribute of spectrocloud_cloudaccount_maas is exercised in one place.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute updates in place.
#
# Prerequisite: a Private Cloud Gateway (PCG) must already be installed and reachable from this
# MAAS environment before this account can be validated.
resource "spectrocloud_cloudaccount_maas" "account" {
  # Required. Display name shown in the Palette UI.
  name = "e2e-maas-account"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required. UID of the Private Cloud Gateway this account routes MAAS traffic through.
  private_cloud_gateway_id = var.private_cloud_gateway_id
  # Required. MAAS API endpoint, e.g. http://maas:5240/MAAS.
  maas_api_endpoint = var.maas_api_endpoint
  # Required, sensitive (credential material). MAAS API key.
  maas_api_key = var.maas_api_key
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_maas" "account" {
#   name = "some-existing-account-name"
# }
