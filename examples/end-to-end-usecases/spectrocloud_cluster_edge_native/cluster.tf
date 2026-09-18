# Comprehensive spectrocloud_cluster_edge_native example - exercises the full attribute surface
# of the resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_edge_native for the minimal happy-path version.
#
# Day-2 mutability summary (see inline comments below for the full breakdown):
#   ForceNew (destroy/recreate): name, the entire cloud_config block.
#   Updates in place: machine_pool (including edge_host assignments), cluster_profile,
#   backup_policy, scan_policy, cluster_rbac_binding, namespaces, host_config, tags, description,
#   and all the top-level Day-2 operational settings.
#
# Note: unlike most other cluster resources, Edge Native has no cloud account concept in
# practice - cloud_account_id is Optional and left unset here, since nodes are provisioned via
# pre-paired edge appliances instead (see edgehost.tf).
resource "spectrocloud_cluster_edge_native" "cluster" {
  # Required, ForceNew.
  name = "e2e-edge-native-cluster"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster exercising the full Edge Native schema."
  # Optional. Arbitrary metadata, e.g. for external tooling to key off of.
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  # Optional, ForceNew. Edge Native clusters are not tied to a cloud provider account, so this is
  # left unset - node placement instead comes from the edge_host assignments below.

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.edge_native_profile.id
    # Supplies values for the profile_variables declared on the profile.
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the edge-native end-to-end usecase example."
    }
    # Optional: override or add a pack's values just for this cluster, without touching the
    # shared profile definition.
    pack {
      name   = "byo-manifest-addon"
      type   = "manifest"
      values = local.byo_manifest_values
    }
  }

  # Alternative to attaching cluster_profile directly: reference a pre-built cluster template
  # (spectrocloud_cluster_config_template) instead. Not normally combined with cluster_profile.
  # cluster_template {
  #   id = spectrocloud_cluster_config_template.template.id
  #   cluster_profile {
  #     id        = spectrocloud_cluster_profile.edge_native_profile.id
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

  # Required, ForceNew (the entire block).
  cloud_config {
    # Optional. Public SSH keys for accessing cluster nodes.
    ssh_keys = var.cluster_ssh_public_keys
    # Optional, Computed. The cluster's virtual IP (address or FQDN). If omitted, Palette
    # assigns one automatically (within overlay_cidr_range, if set).
    # vip = "10.10.232.57"

    # Optional, ForceNew individually. Overlay (VPN) network CIDR carved out for cluster-internal
    # traffic, e.g. across edge sites without a shared L2/L3 network.
    # overlay_cidr_range = "100.64.192.0/23"

    # Optional, default false. Set true for a two-node cluster (no separate, dedicated control
    # plane) - see the `two_node_role` field on edge_host below when this is enabled.
    is_two_node_cluster = false

    # NTP servers used by cluster nodes.
    ntp_servers = ["0.pool.ntp.org", "1.pool.ntp.org", "time.google.com"]
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    # Optional, default "amd64". Allowed: "amd64", "arm64" - matches the paired appliance's CPU
    # architecture.
    arch_type = "amd64"

    # Required, at least one edge_host per machine pool - each maps a pre-paired physical/virtual
    # edge appliance (see edgehost.tf) onto this pool.
    edge_host {
      host_uid  = spectrocloud_appliance.control_plane.uid
      host_name = "e2e-cp-host"
      # Optional networking overrides - if omitted, the appliance keeps its existing network
      # config (e.g. DHCP-assigned address).
      static_ip       = "10.10.32.10"
      default_gateway = "10.10.32.1"
      subnet_mask     = "255.255.255.0"
      dns_servers     = ["10.10.32.2"]
      nic_name        = "eth0"
      # Optional. Only meaningful when cloud_config.is_two_node_cluster = true.
      # two_node_role = "primary"
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

    edge_host {
      host_uid        = spectrocloud_appliance.worker.uid
      host_name       = "e2e-worker-host"
      static_ip       = "10.10.32.11"
      default_gateway = "10.10.32.1"
      subnet_mask     = "255.255.255.0"
      dns_servers     = ["10.10.32.2"]
      nic_name        = "eth0"
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
    prefix             = "e2e-edge-native-backup"
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
