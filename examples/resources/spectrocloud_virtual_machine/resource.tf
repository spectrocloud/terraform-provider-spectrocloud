resource "spectrocloud_virtual_machine" "sivaTestVM" {
  cluster_uid  = "640187ac37918b9611a13919"
  name = "tf-vm-test"
  namespace = "default"
  cpu_cores = 2
  run_on_launch = true
  memory = "5G"
  image = "quay.io/kubevirt/alpine-container-disk-demo"
}