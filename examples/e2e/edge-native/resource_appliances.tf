resource "spectrocloud_appliance" "appliance1" {
  uid = "edgev2-nik-1"
  wait = true
}

resource "spectrocloud_appliance" "appliance2" {
  uid = "edgev2-nik-2"
  wait = true
}