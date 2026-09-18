# Edge Native clusters have no cloud account (cloud_account_id on the cluster resource is
# Optional and unused here) - instead, each machine_pool node maps onto a physical or virtual
# edge appliance that has already been paired with Palette. This registers/manages those
# appliances as resources, exercising every spectrocloud_appliance attribute, rather than looking
# up appliances that were paired out-of-band.
#
# Day-2 mutability: `uid` and `arch_type` are ForceNew (pairing identity/architecture can't
# change post-registration). `tags`, `remote_shell`, and `temporary_shell_credentials` update in
# place.
resource "spectrocloud_appliance" "control_plane" {
  # Required, ForceNew. Must be unique across all appliances (edge hosts) in the tenant, and
  # match the UID the appliance was paired under.
  uid = var.control_plane_edge_host_uid
  # Optional, ForceNew, default "amd64". Allowed: "amd64", "arm64".
  arch_type = "amd64"
  tags = {
    "role" = "control-plane"
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

resource "spectrocloud_appliance" "worker" {
  uid       = var.worker_edge_host_uid
  arch_type = "amd64"
  tags = {
    "role" = "worker"
    "env"  = "e2e-demo"
  }
  wait         = true
  remote_shell = "disabled"
}

# Alternative to registering new appliances: look up ones already paired with Palette.
# data "spectrocloud_appliance" "control_plane" {
#   uid = "some-already-paired-edge-host-uid"
# }
