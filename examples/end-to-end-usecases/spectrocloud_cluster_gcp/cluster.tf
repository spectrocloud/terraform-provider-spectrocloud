# Comprehensive spectrocloud_cluster_gcp example - exercises the full attribute surface of the
# resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_gcp for the minimal happy-path version.
#
# Day-2 mutability summary: only `name` and `cloud_account_id` are ForceNew. Unlike most other
# cluster resources in this provider, `cloud_config` itself is NOT ForceNew here - every field
# in it, including `project` and `region`, updates in place. Everything else - cluster_profile,
# machine_pool, backup_policy, scan_policy, cluster_rbac_binding, namespaces, host_config, and
# the top-level Day-2 settings - also updates in place.
resource "spectrocloud_cluster_gcp" "cluster" {
  # Required, ForceNew.
  name = "e2e-gcp-cluster"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster exercising the full GCP schema."
  # Optional. Arbitrary metadata, e.g. for external tooling to key off of.
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  # Required, ForceNew.
  cloud_account_id = spectrocloud_cloudaccount_gcp.account.id

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.gcp_profile.id
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the gcp end-to-end usecase example."
    }
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
  #     id        = spectrocloud_cluster_profile.gcp_profile.id
  #     variables = { app_version = "1.2.0" }
  #   }
  # }

  pause_agent_upgrades            = "unlock"
  os_patch_on_boot                = false
  os_patch_schedule               = "0 0 * * SUN"
  os_patch_after                  = timeadd(timestamp(), "24h")
  cluster_timezone                = "America/New_York"
  update_worker_pools_in_parallel = false

  # renew_k8s_certificates_now = timestamp()  # trigger, left unset
  # force_delete       = true
  # force_delete_delay = 25
  # skip_completion    = true

  # Not ForceNew - unusual for this provider's cloud_config blocks, but true for GCP.
  cloud_config {
    project = var.gcp_project_id
    region  = var.gcp_region
    # Optional. Name of the VPC network to provision cluster resources into.
    network = var.gcp_network

    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 3
    instance_type           = "n1-standard-4"
    disk_size_gb            = 80
    # Required for this resource (unlike some other clouds where azs is Optional).
    azs = ["us-east1-b", "us-east1-c", "us-east1-d"]
  }

  machine_pool {
    name = "worker-pool"
    additional_labels = {
      "workload" = "general"
    }
    additional_annotations = {
      "team" = "platform"
    }
    taints {
      key    = "dedicated"
      value  = "general"
      effect = "NoSchedule"
    }

    # Note: unlike some other cluster resources, spectrocloud_cluster_gcp's machine_pool has no
    # min/max autoscaling fields - count is the pool's fixed size.
    count = 2

    instance_type = "n1-standard-2"
    disk_size_gb  = 100
    azs           = ["us-east1-b", "us-east1-c"]

    node_repave_interval = 60
    skip_k8s_upgrade     = "disabled"

    update_strategy = "OverrideScaling"
    override_scaling {
      max_surge       = "1"
      max_unavailable = "0"
    }

    override_kubeadm_configuration = <<-EOT
      preKubeadmCommands:
        - echo "preparing worker node"
    EOT

    override_cluster_api_config = <<-EOT
      GCPMachineTemplate:
        spec:
          template:
            spec:
              rootDeviceSize: 100
    EOT

    # node {
    #   node_id = "worker-pool-node-1"
    #   action  = "cordon"
    # }

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
    prefix                         = "e2e-gcp-backup"
    backup_location_id             = spectrocloud_backup_storage_location.bsl.id
    schedule                       = "0 0 * * SUN"
    expiry_in_hour                 = 7200
    include_disks                  = true
    include_cluster_resources_mode = "auto"
    namespaces                     = ["default", "e2e-demo-ns"]
    include_all_clusters           = false
    cluster_uids                   = []
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

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

  host_config {
    host_endpoint_type = "Ingress"
    ingress_host       = "*.e2e-demo.spectrocloud.com"
  }
}
