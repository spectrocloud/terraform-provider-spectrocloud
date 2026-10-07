# Comprehensive spectrocloud_cluster_gke example - exercises the full attribute surface of the
# resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_gke for the minimal happy-path version.
#
# Day-2 mutability summary: `name`, `cloud_account_id`, and both `cloud_config.project`/
# `cloud_config.region` are ForceNew - unlike the plain (non-GKE) spectrocloud_cluster_gcp,
# where cloud_config updates in place. Everything else - cluster_profile, machine_pool,
# backup_policy, scan_policy, cluster_rbac_binding, namespaces, host_config, and the top-level
# Day-2 settings - updates in place.
#
# Note: like AKS/EKS, GKE's control plane is fully Google-managed, so machine_pool has no
# control_plane/control_plane_as_worker fields, and there is no min/max autoscaling or azs
# placement here - GKE's node pool schema is the simplest of this provider's cloud clusters.
resource "spectrocloud_cluster_gke" "cluster" {
  # Required, ForceNew.
  name = "e2e-gke-cluster"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags                   = ["e2e-usecase", "team:platform", "owner:bob"]
  description            = "Comprehensive end-to-end usecase cluster exercising the full GKE schema."
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  cloud_account_id = spectrocloud_cloudaccount_gcp.account.id
  apply_setting    = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.gke_profile.id
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the gke end-to-end usecase example."
    }
    pack {
      name   = "byo-manifest-addon"
      type   = "manifest"
      values = local.byo_manifest_values
    }
  }

  # cluster_template {
  #   id = spectrocloud_cluster_config_template.template.id
  #   cluster_profile {
  #     id        = spectrocloud_cluster_profile.gke_profile.id
  #     variables = { app_version = "1.2.0" }
  #   }
  # }

  pause_agent_upgrades = "unlock"
  os_patch_on_boot     = false
  os_patch_schedule    = "0 0 * * SUN"
  os_patch_after       = timeadd(timestamp(), "24h")
  cluster_timezone     = "America/New_York"

  # Preferred, current field name - "true" updates all worker pools simultaneously.
  update_worker_pools_in_parallel = false
  # Deprecated alias for the field above; the two are mutually exclusive (ConflictsWith). Do not
  # set both.
  # update_worker_pool_in_parallel = false

  # renew_k8s_certificates_now = timestamp()  # trigger, left unset
  # force_delete       = true
  # force_delete_delay = 25
  # skip_completion    = true

  cloud_config {
    # Required, ForceNew.
    project = var.gcp_project_id
    region  = var.gcp_region

    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT
  }

  machine_pool {
    name          = "worker-pool"
    count         = 3
    instance_type = "n1-standard-4"
    # Optional, default 60.
    disk_size_gb = 80

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
  }

  backup_policy {
    prefix                         = "e2e-gke-backup"
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
