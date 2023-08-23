
data "spectrocloud_cluster_profile" "profile1" {
  name = "org-core"
}

data "spectrocloud_cluster_profile" "profile2" {
  name = "spectro-core"
}

resource "spectrocloud_cluster_import" "cluster" {
  name  = "edge-import-tf-14"
  cloud = "generic"
  tags  = ["captain:10.10.01.01", "imported:false"]

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile1.id
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile2.id
  }
}

resource "spectrocloud_appliance" "appliance" {
  uid = "edge-8w7ye9e8e90"
  labels = {
    "name"    = "edge-host-89"
    "cluster" = spectrocloud_cluster_import.cluster.id
  }
}
