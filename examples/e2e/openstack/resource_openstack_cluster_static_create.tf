data "spectrocloud_cloudaccount_openstack" "account" {
  name = var.account_name
}

data "spectrocloud_pack" "csi" {
  name    = var.csi_name
  version = var.csi_ver
}

data "spectrocloud_pack" "cni" {
  name    = var.cni_name
  version = var.cni_ver
}

data "spectrocloud_pack" "k8s" {
  name    = var.k8s_name
  version = var.k8s_ver
}

data "spectrocloud_pack" "ubuntu" {
  name    = var.os_name
  version = var.os_ver
}

resource "spectrocloud_cluster_profile" "profile" {
  name  = var.cp_name
  cloud = var.cloud_name
  type  = "cluster"

  pack {
    name   = data.spectrocloud_pack.ubuntu.name
    tag    = data.spectrocloud_pack.ubuntu.version
    uid    = data.spectrocloud_pack.ubuntu.id
    values = data.spectrocloud_pack.ubuntu.values
  }

  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = data.spectrocloud_pack.csi.name
    tag    = data.spectrocloud_pack.csi.version
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }
}

resource "spectrocloud_cluster_openstack" "cluster" {
  name = var.cluster_name

  cluster_profile {
    id = spectrocloud_cluster_profile.profile.id
  }

  cloud_account_id = data.spectrocloud_cloudaccount_openstack.account.id
  tags             = ["dev"]


  cloud_config {
    domain      = var.domain
    project     = var.project
    region      = var.region
    ssh_key     = var.sshkey
    dns_servers = ["10.10.128.8", "8.8.8.8"]
    subnet_cidr = var.subnet_cidr
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = var.master_inst_count
    instance_type           = var.master_inst_type
    azs                     = ["nova"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = var.worker_inst_count
    instance_type = var.worker_inst_type
    azs           = ["nova"]
  }

}
