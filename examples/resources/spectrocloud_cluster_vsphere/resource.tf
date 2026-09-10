data "spectrocloud_cluster_profile" "vmware_profile" {
  name    = "vsphere-picard-2"
  version = "1.0.0"
  context = "project"
}
data "spectrocloud_cloudaccount_vsphere" "vmware_account" {
  name = var.shared_vmware_cloud_account_name
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

# Day-2 mutability: `name` and `cloud_account_id` are ForceNew. The entire `cloud_config` block
# is ForceNew - changing any field inside it (datacenter, folder, ssh_keys, network_type, etc.)
# recreates the cluster. `machine_pool.name`, `control_plane`, and `control_plane_as_worker` are
# NOT ForceNew for this resource (they update in place), unlike the equivalent fields on some
# other cloud resources in this provider. `cluster_profile`, `backup_policy`, `scan_policy`,
# `tags`, and the rest of `machine_pool` also update in place.
resource "spectrocloud_cluster_vsphere" "cluster" {
  name = "vsphere-picard-3"
  # Optional, default false/20. Force-delete the cluster without waiting for the provisioned
  # cloud resources to clean up first - you are then responsible for cleaning them up manually.
  # force_delete_delay only takes effect when force_delete is true; minimum is 20 minutes.
  # force_delete = true
  # force_delete_delay = 25
  cloud_account_id = data.spectrocloud_cloudaccount_vsphere.vmware_account.id
  cluster_profile {
    id = data.spectrocloud_cluster_profile.vmware_profile.id
  }
  cloud_config {
    # ssh_keys is preferred over the deprecated singular ssh_key field.
    ssh_keys = [var.cluster_ssh_public_key]

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder
    // For Dynamic DNS (network_type & network_search_domain value should set for DDNS)
    network_type          = "DDNS"
    network_search_domain = var.cluster_network_search
    // For Static (By Default static_ip is false, for static provisioning, it is set to be true. Not required to specify network_type & network_search_domain)
    # static_ip = true

    # Optional: YAML passthrough for CAPV properties not yet first-class in
    # Palette. Overrides pack-level and Palette-managed values. Palette does
    # not pre-validate keys/types/values; the API surfaces any errors.
    # override_cluster_api_config = <<-EOT
    #   VSphereCluster:
    #     spec:
    #       identityRef:
    #         kind: VSphereClusterIdentity
    #         name: my-identity
    # EOT
  }

  backup_policy {
    schedule                  = "0 0 * * SUN"
    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
    prefix                    = "prod-backup"
    expiry_in_hour            = 7200
    include_disks             = true
    include_cluster_resources = true
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1
    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
      # Optional. Required only when cloud_config.static_ip is true.
      # static_ip_pool_id = data.spectrocloud_ippool.pool.id
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 4096
      cpu          = 2
    }
  }

  machine_pool {
    name                 = "worker-basic"
    count                = 1
    node_repave_interval = 30
    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 40
      memory_mb    = 8192
      cpu          = 4
    }

    # Optional, default "disabled". "enabled" decouples this worker pool's Kubernetes upgrade
    # from the control plane (up to N-3 minor-version skew); applicable only to worker pools.
    # skip_k8s_upgrade = "disabled"

    # Optional: YAML passthrough for pool-level CAPV properties (e.g. VSphereMachineTemplate).
    # override_cluster_api_config = <<-EOT
    #   VSphereMachineTemplate:
    #     spec:
    #       template:
    #         spec:
    #           diskGiB: 80
    # EOT

    # Optional: override Machine Health Check settings for this node pool
    override_health_check_configuration = <<-EOT
      maxUnhealthy: 40%
      nodeStartupTimeout: 10m
      unhealthyConditions:
        - type: Ready
          status: "False"
          timeout: 5m
        - type: Ready
          status: "Unknown"
          timeout: 5m
    EOT
  }
}

# Other optional, less commonly changed top-level attributes not shown above: cluster_meta_attribute,
# apply_setting, review_repave_state, pause_agent_upgrades, os_patch_on_boot/os_patch_schedule/os_patch_after,
# cluster_timezone, update_worker_pools_in_parallel, skip_completion, host_config, location_config,
# namespaces, cluster_rbac_binding, cluster_template. None of these are ForceNew.