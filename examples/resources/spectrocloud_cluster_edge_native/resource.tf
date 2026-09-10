# Day-2 mutability: `name` and `cloud_account_id` are ForceNew. The entire `cloud_config` block
# is also ForceNew - changing ssh_keys, vip, overlay_cidr_range, etc. recreates the cluster.
# `cluster_profile` and `machine_pool` (including edge_host assignments) update in place.
resource "spectrocloud_cluster_edge_native" "cluster" {
  name = "ran-edge-tf"

  cluster_profile {
    id = "test-profile-id"
  }

  cloud_config {
    # Optional. Public SSH keys for accessing cluster nodes.
    ssh_keys = ["spectro2023"]
    # Optional, Computed. The cluster's virtual IP - an address or FQDN. If omitted, Palette
    # assigns one automatically (within overlay_cidr_range, if set).
    vip = "10.10.232.57"
    # Optional. Overlay (VPN) network CIDR, e.g. "100.64.192.0/23". Also individually ForceNew.
    # overlay_cidr_range = "100.64.192.0/23"
    # Optional, default false. Set to true for a two-node (no separate control plane) cluster.
    # is_two_node_cluster = false
    # Optional. NTP servers for the cluster to use.
    # ntp_servers = ["pool.ntp.org"]
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    # Optional, default "amd64". Allowed: "amd64", "arm64".
    arch_type = "amd64"

    # Required, at least one edge_host per machine pool - each maps a physical/virtual edge
    # appliance (already paired to Palette) onto this pool.
    edge_host {
      # Required. UID of the paired edge appliance (see spectrocloud_appliance).
      host_uid = "edge-fsdsdedadfasdtest"
      # Optional networking overrides - if omitted, the appliance keeps its existing network
      # config (e.g. DHCP-assigned address).
      static_ip       = "10.10.32.12"
      default_gateway = "10.10.12.1"
      dns_servers     = ["tf.test.com"]
      host_name       = "test-test"
      nic_name        = "auto162"
      subnet_mask     = "255.255.12.0"
      # Optional. Only for is_two_node_cluster = true. Allowed: "primary", "secondary".
      # two_node_role = "primary"
    }
  }

  machine_pool {
    name = "wp-pool"
    # Optional, default "disabled". "enabled" skips the OS/K8s upgrade for this worker pool
    # (N-3 skew allowed) when the cluster profile is upgraded.
    skip_k8s_upgrade = "disabled"

    edge_host {
      host_uid        = "edge-bef8384adfasdtest"
      default_gateway = "10.10.12.1"
      dns_servers     = ["tf.test.com"]
      host_name       = "test-test"
      nic_name        = "auto160"
      static_ip       = "10.10.44.22"
      subnet_mask     = "255.255.92.0"
    }
  }
}
