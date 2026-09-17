# Comprehensive spectrocloud_virtual_cluster example - exercises the full attribute surface of
# the resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_virtual_cluster for the minimal happy-path version.
#
# Unlike every other cluster resource in this repo, a virtual cluster has no cloud account and
# provisions no infrastructure of its own - it's a Kubernetes-in-Kubernetes control plane
# (vcluster) that runs inside an already-existing Palette host cluster.
#
# Day-2 mutability: only `name` and `cloud_config` are ForceNew - changing either recreates the
# virtual cluster. Everything else - host_cluster_uid, cluster_group_uid, resources,
# cluster_profile, tags, description, backup_policy, scan_policy, cluster_rbac_binding,
# namespaces, and the top-level Day-2 settings - updates in place.
resource "spectrocloud_virtual_cluster" "cluster" {
  # Required, ForceNew.
  name = "e2e-virtual-cluster"
  # Optional, default "project". Despite the schema description mentioning "tenant", the only
  # values actually accepted are "project" or "cluster".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase virtual cluster exercising the full schema."

  # Required in practice: set exactly one of host_cluster_uid or cluster_group_uid to place the
  # virtual cluster - directly on a host cluster (this example), or on whichever cluster a
  # cluster group selects.
  host_cluster_uid = var.host_cluster_uid
  # Alternative to a specific host cluster: let a cluster group pick the placement instead.
  # cluster_group_uid = data.spectrocloud_cluster_group.cg.id

  # Optional, default false. Set true to pause the running cluster (scales the vcluster control
  # plane to zero); false to resume it.
  pause_cluster = false

  # Optional. Resource quota bounds enforced on this virtual cluster within its host - all 6
  # fields optional, set only the ones you want to bound.
  resources {
    max_cpu           = 6
    max_mem_in_mb     = 6000
    max_storage_in_gb = 20
    min_cpu           = 0
    min_mem_in_mb     = 0
    min_storage_in_gb = 0
  }

  cluster_profile {
    id = spectrocloud_cluster_profile.virtual_cluster_profile.id
    # Supplies values for the profile_variables declared on the profile.
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the virtual-cluster end-to-end usecase example."
    }
    # Optional: override or add a pack's values just for this cluster, without touching the
    # shared profile definition.
    pack {
      name   = "byo-manifest-addon"
      type   = "manifest"
      values = local.byo_manifest_values
    }
  }

  # Optional, default "unlock". "lock" pauses automatic Palette agent/component upgrades.
  pause_agent_upgrades = "unlock"

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  # Optional, default false. true updates all worker pools simultaneously - not very meaningful
  # for a single-control-plane virtual cluster, but exercised here for completeness.
  update_worker_pools_in_parallel = false

  # Optional, default "". Must be an IANA timezone.
  cluster_timezone = "America/New_York"

  # Optional. Set to the current time to trigger an immediate control-plane cert renewal on this
  # apply - it is a trigger, not a schedule, so it is left unset here.
  # renew_k8s_certificates_now = timestamp()

  # Optional, default false. Applies queued OS patches automatically on node boot.
  os_patch_on_boot = false
  # Optional. Cron schedule for OS patching.
  os_patch_schedule = "0 0 * * SUN"
  # Optional. Dynamic so this example never ships with a stale, already-past timestamp.
  os_patch_after = timeadd(timestamp(), "24h")

  # Optional, default false/20. Force-delete without waiting for the vcluster to clean up first;
  # force_delete_delay only applies when force_delete is true (minimum 20 minutes).
  # force_delete       = true
  # force_delete_delay = 25

  # Optional, default false. Create asynchronously without waiting for provisioning to finish.
  # skip_completion = true

  # Optional, ForceNew (the entire block, if set). Configures the Helm chart used to install the
  # virtual cluster's control plane - if omitted, Palette uses its own default vcluster chart.
  cloud_config {
    chart_name    = "vcluster"
    chart_repo    = "https://charts.loft.sh"
    chart_version = "0.19.5"
    k8s_version   = "1.28.5"
    chart_values  = <<-EOT
      syncer:
        extraArgs:
          - --tls-san=e2e-virtual-cluster.example.com
    EOT
  }

  backup_policy {
    prefix             = "e2e-virtual-cluster-backup"
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
}
