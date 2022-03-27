resource "spectrocloud_appliance" "appliance" {
  uid = "nik-libvirt15-mar-20"
  labels = {
    "name" = "nik_appliance_name"
  }
  wait = true
}