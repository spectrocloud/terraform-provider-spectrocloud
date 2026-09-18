# Comprehensive spectrocloud_cluster_vsphere example - exercises the full attribute surface of
# the resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_vsphere for the minimal happy-path version.
#
# Day-2 mutability summary (see inline comments below for the full breakdown):
#   ForceNew (destroy/recreate): name, cloud_account_id, the entire cloud_config block.
#   Updates in place: everything else - cluster_profile, machine_pool, backup_policy,
#   scan_policy, cluster_rbac_binding, namespaces, host_config, tags, description, and all the
#   top-level Day-2 operational settings.
resource "spectrocloud_cluster_vsphere" "cluster" {
  # Required, ForceNew.
  name = "e2e-vsphere-cluster"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster exercising the full vSphere schema."
  # Optional. Arbitrary metadata, e.g. for external tooling to key off of.
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  # Required, ForceNew. References the cloud account created in cloudaccount.tf.
  cloud_account_id = spectrocloud_cloudaccount_vsphere.account.id

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.vsphere_profile.id
    # Supplies values for the profile_variables declared on the profile.
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the vsphere end-to-end usecase example."
    }
    # Optional: override or add a pack's values just for this cluster, without touching the
    # shared profile definition.
    pack {
      name   = "spectro-proxy-addon"
      type   = "manifest"
      values = local.proxy_addon_values
    }
  }

  # Alternative to attaching cluster_profile directly: reference a pre-built cluster template
  # (spectrocloud_cluster_config_template) instead. Not normally combined with cluster_profile.
  # cluster_template {
  #   id = spectrocloud_cluster_config_template.template.id
  #   cluster_profile {
  #     id        = spectrocloud_cluster_profile.vsphere_profile.id
  #     variables = { app_version = "1.2.0" }
  #   }
  # }

  # Optional, default "unlock". "lock" pauses automatic Palette agent/component upgrades.
  pause_agent_upgrades = "unlock"

  # Optional, default false. Applies queued OS patches automatically on node boot.
  os_patch_on_boot = false
  # Optional. Cron schedule for OS patching.
  os_patch_schedule = "0 0 * * SUN"
  # Optional. Dynamic so this example never ships with a stale, already-past timestamp.
  os_patch_after = timeadd(timestamp(), "24h")

  # Optional, default "". Must be an IANA timezone.
  cluster_timezone = "America/New_York"

  # Optional, default false. true updates all worker pools simultaneously.
  update_worker_pools_in_parallel = false

  # Optional. Set to the current time to trigger an immediate control-plane cert renewal on this
  # apply - it is a trigger, not a schedule, so it is left unset here.
  # renew_k8s_certificates_now = timestamp()

  # Optional, default false/20. Force-delete without waiting for cloud resources to clean up
  # first; force_delete_delay only applies when force_delete is true (minimum 20 minutes).
  # force_delete       = true
  # force_delete_delay = 25

  # Optional, default false. Create asynchronously without waiting for provisioning to finish.
  # skip_completion = true

  # Required, ForceNew (the entire block). vSphere placement and network configuration.
  cloud_config {
    datacenter = var.vsphere_datacenter
    folder     = var.vsphere_folder
    # Optional. vSphere folder holding VM image templates. Defaults to "spectro-templates".
    image_template_folder = "spectro-templates"

    # ssh_keys is preferred over the deprecated singular ssh_key field.
    ssh_keys = [var.cluster_ssh_public_key]

    # Networking: DDNS (shown active) and static IP are mutually exclusive.
    static_ip = false
    # For Dynamic DNS - network_type and network_search_domain apply when static_ip = false.
    network_type          = "DDNS"
    network_search_domain = "corp.example.com"
    # host_endpoint controls whether the cluster's endpoint is exposed as an IP or an
    # FQDN (external/DDNS) - relevant when static_ip = false.
    host_endpoint = "FQDN"

    # For static provisioning instead (mutually exclusive with the DDNS settings above):
    # static_ip = true
    # (leave network_type/network_search_domain unset, and set placement.static_ip_pool_id
    # on each machine_pool below)

    # NTP servers used by cluster nodes.
    ntp_servers = ["0.pool.ntp.org", "1.pool.ntp.org", "time.google.com"]

    # Optional: YAML passthrough for CAPV properties not yet first-class in Palette. Overrides
    # pack-level and Palette-managed values; not ForceNew despite the rest of cloud_config being
    # ForceNew as a whole.
    override_cluster_api_config = <<-EOT
      VSphereCluster:
        spec:
          identityRef:
            kind: VSphereClusterIdentity
            name: my-identity
    EOT
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 3
    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
    }
    instance_type {
      disk_size_gb = 60
      memory_mb    = 8192
      cpu          = 4
    }
  }

  machine_pool {
    name = "worker-pool"
    # Optional. Additional Kubernetes labels/annotations applied to every node in this pool.
    additional_labels = {
      "workload" = "general"
    }
    additional_annotations = {
      "team" = "platform"
    }
    # Optional. Taints repel pods that don't tolerate them.
    taints {
      key    = "dedicated"
      value  = "general"
      effect = "NoSchedule"
    }

    # Autoscaling: set count to the initial size, min/max to bound the autoscaler's range.
    count = 2
    min   = 2
    max   = 6

    # Optional, default 0. Minimum seconds a node must stay Ready before the next is repaved.
    # Applicable only to worker pools.
    node_repave_interval = 60
    # Optional, default "disabled". "enabled" decouples this pool's K8s upgrade from the control
    # plane (up to N-3 minor-version skew) - useful for staged rollouts.
    skip_k8s_upgrade = "disabled"

    # Optional, default "RollingUpdateScaleOut". "OverrideScaling" requires override_scaling
    # below with both max_surge and max_unavailable set.
    update_strategy = "OverrideScaling"
    override_scaling {
      max_surge       = "1"
      max_unavailable = "0"
    }

    # Optional: YAML for kubeletExtraArgs/pre/post kubeadm commands. Worker pools only.
    override_kubeadm_configuration = <<-EOT
      preKubeadmCommands:
        - echo "preparing worker node"
    EOT

    # Optional: YAML passthrough for pool-level CAPV properties.
    override_cluster_api_config = <<-EOT
      VSphereMachineTemplate:
        spec:
          template:
            spec:
              diskGiB: 80
    EOT

    placement {
      cluster       = var.vsphere_cluster
      resource_pool = var.vsphere_resource_pool
      datastore     = var.vsphere_datastore
      network       = var.vsphere_network
      # Required only when cloud_config.static_ip is true.
      # static_ip_pool_id = data.spectrocloud_ippool.pool.id
    }
    instance_type {
      disk_size_gb = 100
      memory_mb    = 16384
      cpu          = 8
    }

    # Optional: cordon/uncordon a specific node in this pool (Day-2 action).
    # node {
    #   node_id = "worker-pool-node-1"
    #   action  = "cordon"
    # }

    # Optional: override Machine Health Check settings for this node pool.
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

  backup_policy {
    prefix             = "e2e-vsphere-backup"
    backup_location_id = spectrocloud_backup_storage_location.bsl.id
    schedule           = "0 0 * * SUN"
    expiry_in_hour     = 7200
    include_disks      = true
    # include_cluster_resources and include_cluster_resources_mode are mutually exclusive -
    # using the newer mode-based field here.
    include_cluster_resources_mode = "auto"
    namespaces                     = ["default", "e2e-demo-ns"]
    include_all_clusters           = false
    # cluster_uids is only meaningful when include_all_clusters = false, to back up specific
    # other clusters alongside this one - left empty since this policy is for this cluster only.
    cluster_uids = []
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  # Cluster-scoped RBAC binding.
  cluster_rbac_binding {
    type = "ClusterRoleBinding"
    role = {
      kind = "ClusterRole"
      name = "cluster-admin"
    }
    subjects {
      type = "User"
      name = "e2e-cluster-admin-user"
    }
    subjects {
      type = "Group"
      name = "e2e-cluster-admins"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "e2e-admin-sa"
      namespace = "kube-system"
    }
  }

  # Namespace-scoped RBAC binding - namespace is required when type = "RoleBinding".
  cluster_rbac_binding {
    type      = "RoleBinding"
    namespace = "e2e-demo-ns"
    role = {
      kind = "Role"
      name = "e2e-demo-editor"
    }
    subjects {
      type = "User"
      name = "e2e-demo-user"
    }
  }

  namespaces {
    name = "e2e-demo-ns"
    resource_allocation = {
      cpu_cores    = "4"
      memory_MiB   = "4096"
      gpu_limit    = "0"
      gpu_provider = "none"
    }
  }

  # Optional: expose the cluster via Ingress (shown) or LoadBalancer (commented alternative).
  host_config {
    host_endpoint_type = "Ingress"
    ingress_host       = "*.e2e-demo.spectrocloud.com"
  }
  # host_config {
  #   host_endpoint_type           = "LoadBalancer"
  #   external_traffic_policy      = "Local"
  #   load_balancer_source_ranges  = "10.0.0.0/8"
  # }

  # location_config is Computed-only for this resource (Palette derives it) - see outputs.tf.
}
