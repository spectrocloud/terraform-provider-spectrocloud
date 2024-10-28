resource "spectrocloud_cluster_edge_native" "cluster" {
  name = "ran-edge-tf"

  cluster_profile {
    id = "test-profile-id"
  }

  cloud_config {
    ssh_keys = ["spectro2023"]
    vip      = "10.10.232.57"
    #    overlay_cidr_range = "100.64.192.0/23"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"

    edge_host {
      host_uid        = "edge-fsdsdedadfasdtest"
      static_ip       = "10.10.32.12"
      default_gateway = "10.10.12.1"
      dns_servers     = ["tf.test.com"]
      host_name       = "test-test"
      nic_name        = "auto162"
      subnet_mask     = "255.255.12.0"
    }
  }

  machine_pool {
    name = "wp-pool"

    edge_host {
      host_uid        = "edge-bef8384adfasdtest"
      default_gateway = "10.10.12.1"
      dns_servers     = ["tf.test.com"]
      host_name       = "test-test"
      nic_name        = "auto160"
      static_ip       = "10.10.132449.22"
      subnet_mask     = "255.255.92.0"
    }
  }
}
