# Attributes:
#   uid                         - Required, ForceNew. Must be unique across all appliances (edge
#                                 hosts) in the tenant.
#   arch_type                   - Optional, ForceNew, default "amd64". Allowed: "amd64", "arm64".
#   tags                        - Optional. Arbitrary key-value labels for organizing appliances.
#   pairing_key                 - Optional. The one-time pairing key shown when registering the
#                                 edge host in Palette. Only needed the first time this appliance
#                                 is paired - omit on subsequent applies.
#   wait                        - Optional, ForceNew, defaults to false. If true, Terraform waits
#                                 for the appliance to reach "ready"/"healthy" state during create
#                                 before returning.
#   remote_shell                - Optional, default "disabled". Allowed: "enabled"/"disabled".
#                                 Enables SSH remote-shell access to the edge host from Palette.
#                                 Updatable in place.
#   temporary_shell_credentials - Optional, default "disabled". Allowed: "enabled"/"disabled".
#                                 Creates a temporary sudo user on the edge host for SSH
#                                 auto-login; deleted again on deactivation. Can only be "enabled"
#                                 when remote_shell is also "enabled" - the provider rejects the
#                                 opposite combination.
resource "spectrocloud_appliance" "appliance" {
  uid       = "test-dec9"
  arch_type = "amd64"

  tags = {
    "name" = "appliance_name"
  }

  # pairing_key = "REPLACE_ME"

  wait                        = true
  remote_shell                = "disabled"
  temporary_shell_credentials = "disabled"
}
