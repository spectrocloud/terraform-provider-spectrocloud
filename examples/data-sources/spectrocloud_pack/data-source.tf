data "spectrocloud_pack" "k8sgpt" {
  name = "k8sgpt-operator"
  advance_filters {
    pack_type   = ["spectro"]
    addon_type  = ["system app"]
    pack_layer  = ["addon"]
    environment = ["all"]
    is_fips     = false
    pack_source = ["community"]
  }
  registry_uid = "5ee9c5adc172449eeb9c30cf"
  #  version = "3.16.0"
  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
  # cloud = ["azure"]
}

output "same" {
  value = data.spectrocloud_pack.k8sgpt
}
