resource "virtual_machine" "example" {
  cluster_uid = spectrocloud_cluster_vsphere.cluster.id
  name        = var.name
  namespace   = var.namespace

  /*template {
    id = var.template_uid
  }*/
  state               = var.state
  cpu_cores           = var.cpu_cores
  memory              = var.memory
  image               = var.image
  cloudinit_user_data = var.cloudinit_user_data
}