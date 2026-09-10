# Day-2 mutability: only `name` and `cloud_config` are ForceNew - changing either recreates the
# virtual cluster. Everything else - host_cluster_uid, cluster_group_uid, resources,
# cluster_profile, tags, description, and the rest - updates in place.
resource "spectrocloud_virtual_cluster" "cluster" {
  # Required, ForceNew.
  name = "virtual-cluster-demo"

  # Optional, default "project". Despite the schema description mentioning `tenant`, the only
  # values actually accepted are "project" or "cluster".
  # context = "project"

  # Optional. Tags in `key:value` form.
  # tags = ["dev", "department:devops", "owner:bob"]

  # Optional, default "". Free-text description.
  # description = "Demo virtual cluster"

  # Required in practice: set exactly one of host_cluster_uid or cluster_group_uid to place the
  # virtual cluster - directly on a host cluster (this example), or on whichever cluster a
  # cluster group selects.
  host_cluster_uid = var.host_cluster_uid
  # cluster_group_uid = data.spectrocloud_cluster_group.cg.id

  # Optional, default false. Set true to pause the cluster; false to resume it.
  # pause_cluster = false

  resources {
    # All 6 fields optional; set only the ones you want to bound.
    max_cpu           = 6
    max_mem_in_mb     = 6000
    max_storage_in_gb = 20
    min_cpu           = 0
    min_mem_in_mb     = 0
    min_storage_in_gb = 0
  }
}

# Other optional, less commonly changed attributes not shown above: cluster_profile,
# pause_agent_upgrades, apply_setting, update_worker_pools_in_parallel, cluster_timezone,
# os_patch_on_boot/os_patch_schedule/os_patch_after, backup_policy, scan_policy,
# cluster_rbac_binding, namespaces, force_delete/force_delete_delay. None of these are ForceNew.
# `cloud_config` (chart_name/chart_repo/chart_version/chart_values) IS ForceNew if you do set it -
# it configures the Helm chart used to install the virtual cluster's control plane.