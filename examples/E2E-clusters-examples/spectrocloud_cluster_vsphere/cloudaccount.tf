# Creates and manages a dedicated vSphere cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained and every
# attribute of spectrocloud_cloudaccount_vsphere is exercised in one place.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute, including the
# vCenter credentials, updates in place.
#
# Prerequisite: a Private Cloud Gateway (PCG) must already be installed and reachable from this
# vCenter before this account can be validated. See
# https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster
resource "spectrocloud_cloudaccount_vsphere" "account" {
  # Required. Display name shown in the Palette UI.
  name = "e2e-vsphere-account"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required. UID of the Private Cloud Gateway this account routes vCenter traffic through.
  private_cloud_gateway_id = var.private_cloud_gateway_id
  # Required. vCenter server address (hostname or IP), e.g. "vcenter.corp.example.com".
  vsphere_vcenter = var.vsphere_vcenter
  # Required. vCenter username with permissions to provision/manage clusters.
  vsphere_username = var.vsphere_username
  # Required, sensitive (credential material). vCenter password.
  vsphere_password = var.vsphere_password
  # Optional, default false. Set true to accept vCenter's certificate even if it's self-signed
  # or otherwise fails standard TLS validation - only use this for trusted internal networks.
  vsphere_ignore_insecure_error = var.vsphere_ignore_insecure_error
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_vsphere" "account" {
#   name = "some-existing-account-name"
# }
