data "spectrocloud_pack" "cni-calico" {
  name    = "cni-calico"
  version = "3.16.0"

  # (alternatively)
  # id =  "5fd0ca727c411c71b55a359c"
  # name = "cni-calico-azure"
  # cloud = ["azure"]
}

output "same" {
  value = data.spectrocloud_pack.cni-calico
}
