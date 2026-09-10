# Edge vSphere clusters have no cloud account - instead, the whole cluster is deployed onto a
# single physical or virtual edge appliance (running a vSphere-compatible virtualization layer)
# that has already been paired with Palette. This registers/manages that appliance as a
# resource, exercising every spectrocloud_appliance attribute, rather than looking up an
# appliance that was paired out-of-band.
#
# Day-2 mutability: `uid` and `arch_type` are ForceNew (pairing identity/architecture can't
# change post-registration). `tags`, `remote_shell`, and `temporary_shell_credentials` update in
# place.
resource "spectrocloud_appliance" "edge_host" {
  # Required, ForceNew. Must be unique across all appliances (edge hosts) in the tenant, and
  # match the UID the appliance was paired under.
  uid = var.edge_host_uid
  # Optional, ForceNew, default "amd64". Allowed: "amd64", "arm64".
  arch_type = "amd64"
  tags = {
    "role" = "edge-vsphere-host"
    "env"  = "e2e-demo"
  }
  # Optional. Only needed the first time this appliance is paired - the one-time pairing key
  # shown when registering the edge host in Palette. Omit on subsequent applies.
  # pairing_key = "REPLACE_ME"
  # Optional, ForceNew, default false. Waits for the appliance to reach "ready"/"healthy" state
  # during create before returning.
  wait = true
  # Optional, default "disabled". Enables SSH remote-shell access to this edge host from Palette.
  remote_shell = "disabled"
  # Optional, default "disabled". Only valid when remote_shell = "enabled" - creates a temporary
  # sudo user on the host for SSH auto-login.
  temporary_shell_credentials = "disabled"
}

# Alternative to registering a new appliance: look up one already paired with Palette.
# data "spectrocloud_appliance" "edge_host" {
#   uid = "some-already-paired-edge-host-uid"
# }
