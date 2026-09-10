resource "spectrocloud_appliance" "appliance" {
  # Required, ForceNew. Must be unique across all appliances (edge hosts) in the tenant.
  uid = "test-dec9"

  # Optional, ForceNew, default "amd64". Allowed values: "amd64", "arm64".
  arch_type = "amd64"

  # Optional. Arbitrary key-value labels for organizing appliances.
  tags = {
    "name" = "appliance_name"
  }

  # Optional. The one-time pairing key shown when registering the edge host in Palette. Only
  # needed the first time this appliance is paired - omit on subsequent applies.
  # pairing_key = "REPLACE_ME"

  # Optional, ForceNew, defaults to false. If true, Terraform waits for the appliance to reach
  # "ready"/"healthy" state during create before returning.
  wait = true

  # Optional, default "disabled". Allowed: "enabled"/"disabled". Enables SSH remote-shell access
  # to the edge host from Palette. Updatable in place.
  remote_shell = "disabled"

  # Optional, default "disabled". Allowed: "enabled"/"disabled". Creates a temporary sudo user on
  # the edge host for SSH auto-login; deleted again on deactivation. Can only be "enabled" when
  # remote_shell is also "enabled" - the provider rejects the opposite combination.
  temporary_shell_credentials = "disabled"
}
