data "spectrocloud_cluster_profile" "vmware_profile" {
  name    = "edge-vsphere-picard"
  version = "1.0.0"
  context = "project"
}

data "spectrocloud_backup_storage_location" "bsl" {
  name = var.backup_storage_location_name
}

# Day-2 mutability: `name`, `edge_host_uid`, and the entire `cloud_config` block are ForceNew -
# changing any of them recreates the cluster. `cluster_profile`, `backup_policy`, `scan_policy`,
# `tags`, and `machine_pool` all update in place.
resource "spectrocloud_cluster_edge_vsphere" "cluster" {
  # Required, ForceNew.
  name = "edge-vsphere-picard-1"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Required, ForceNew. UID of the Edge host (spectrocloud_appliance) this cluster runs on.
  edge_host_uid = var.edge_host_uid

  # Optional, default false/20. Force-delete the cluster without waiting for the provisioned
  # cloud resources to clean up first - you are then responsible for cleaning them up manually.
  # force_delete_delay only takes effect when force_delete is true; minimum is 20 minutes.
  # force_delete = true
  # force_delete_delay = 25

  cluster_profile {
    id = data.spectrocloud_cluster_profile.vmware_profile.id
  }

  cloud_config {
    # ssh_keys is preferred over the deprecated singular ssh_key field.
    ssh_keys = [var.cluster_ssh_public_key]

    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder
    # Optional. vSphere folder holding VM image templates for node provisioning. Defaults to
    # "spectro-templates" when unset.
    # image_template_folder = "spectro-templates"

    # Required. Virtual IP for the Kubernetes control plane endpoint.
    vip = var.cluster_vip
    # Optional, default false. Use a static IP instead of DHCP for the control plane endpoint.
    # static_ip = true
    # Optional. Network endpoint type for the control plane address (e.g. "DDNS").
    # network_type = "DDNS"
    # Optional. DNS search domain for the control plane endpoint, used with DDNS.
    # network_search_domain = "corp.example.com"
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

    # Optional: cordon/uncordon a specific node in this pool (Day-2 action, not applied at
    # create time).
    # node {
    #   node_id = "i-07f899a33dee624f7"
    #   action  = "cordon"
    # }

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

# Other optional, less commonly changed top-level attributes not shown above: tags, description,
# cluster_meta_attribute, cluster_template, review_repave_state, pause_agent_upgrades,
# os_patch_on_boot/os_patch_schedule/os_patch_after, cluster_timezone,
# update_worker_pools_in_parallel, skip_completion, cluster_rbac_binding, namespaces,
# host_config, location_config (settable country_code/latitude/longitude). None of these are
# ForceNew.
