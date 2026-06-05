resource "spectrocloud_appliance" "appliance0" {
  uid       = "edgev2-nik-0"
  arch_type = "amd64"
  wait      = true
}

resource "spectrocloud_appliance" "appliance1" {
  uid       = "edgev2-nik-1"
  arch_type = "arm64"
  wait      = true
}