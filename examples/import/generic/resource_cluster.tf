
data "spectrocloud_cluster_profile" "profile1" {
  name = "edge-ubuntu-k3s"
}

data "spectrocloud_cluster_profile" "profile2" {
  name = "spectro-core"
}

resource "spectrocloud_cluster_import" "cluster" {
  name               = "edge-p6os-ha1"
  cloud              = "generic"
  tags        = ["imported:false"]

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile1.id
  }

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile2.id
  }
}

resource "spectrocloud_appliance" "appliance" {
  uid = "edge-d4043673d92f"
  labels = {
    "name" = "edge-host-92"
    "cluster" = spectrocloud_cluster_import.cluster.id
  }

  depends_on = [local_file.appliance]

  provisioner "local-exec" {
    when    = destroy
    command = format("chmod +x delete_%s.sh; bash delete_%s.sh;", spectrocloud_cluster_import.cluster.id, spectrocloud_cluster_import.cluster.id)
  }
}

data "template_file" "delete_script" {
  template = file("delete_cluster.sh")
  vars = {
    API_KEY = var.sc_api_key
    SC_HOST = var.sc_host
    PROJECT_ID = var.sc_project_id
    CLUSTER_ID = spectrocloud_cluster_import.cluster.id
  }
}

resource "local_file" "appliance" {
  content  = data.template_file.delete_script.rendered
  filename = format("delete_%s.sh", spectrocloud_cluster_import.cluster.id)
}


