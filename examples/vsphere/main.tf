terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud.com/spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host        = "console.spectrocloud.com"
  username    = "<....>"
  password    = "<....>"
  project_uid = "<....>"
}

resource "spectrocloud_cloudaccount_vsphere" "vsphere-1" {
  name = "vsphere-1"
  private_cloud_gateway_id = "<....>"
  vsphere_vcenter = "<....>"
  vsphere_username = "<....>"
  vsphere_password = "<....>"
  vsphere_ignore_insecure_error = true
}

resource "spectrocloud_cluster_vsphere" "test6" {
  name               = "test6"
  cluster_profile_id = "<....>"
  cloud_account_id   = spectrocloud_cloudaccount_vsphere.vsphere-1.id

  cloud_config {
    # Replace with your own
    datacenter = "Datacenter"
    folder = "Demo/spc-test4"

    network_type = "DDNS"
    # Replace
    network_search_domain = "spectrocloud.local"

    # Replace
    ssh_key         = "ssh-rsa AAA...."

  }

  # To override values
  #
  # pack {
  #   name   = "spectro-byo-manifest"
  #   tag    = "1.0.x"
  #   values = <<-EOT
  #     manifests:
  #       byo-manifest:
  #         contents: |
  #           # Add manifests here
  #           apiVersion: v1
  #           kind: Namespace
  #           metadata:
  #             labels:
  #               app: wordpress
  #               app2: wordpress2
  #             name: wordpress
  #   EOT
  # }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1

    placement {
      cluster = "cluster1"
      resource_pool = ""
      datastore = "datastore55"
      network = "VM Network"
    }
    instance_type {
      disk_size_gb = 61
      memory_mb = 4096
      cpu = 2
    }
  }

  machine_pool {
    name          = "worker-basic"
    count         = 2

    placement {
      cluster = "cluster1"
      resource_pool = ""
      datastore = "datastore55"
      network = "VM Network"
    }
    instance_type {
      disk_size_gb = 65
      memory_mb = 8192
      cpu = 4
    }
  }


}
