data "spectrocloud_cloudaccount_gcp" "account" {
  name = var.gcp_cloud_account_name
}

data "spectrocloud_cluster_profile" "profile" {
  name = var.gke_cluster_profile_name
}


# Day-2 mutability: `name` and `cloud_account_id` are ForceNew, and so are `cloud_config.project`
# and `cloud_config.region` - unlike the plain (non-GKE) GCP cluster resource, where those two
# update in place. `cluster_profile`, `machine_pool`, and the worker-pool-parallel-update setting
# below all update in place.
resource "spectrocloud_cluster_gke" "cluster" {
  name             = var.cluster_name
  description      = "Gke Cluster"
  tags             = ["dev", "department:pax"]
  cloud_account_id = data.spectrocloud_cloudaccount_gcp.account.id
  context          = "project"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    project = var.gcp_project
    region  = var.gcp_region

    # Optional: YAML passthrough for CAPG-managed (GKE) properties not yet
    # first-class in Palette. Overrides pack-level and Palette-managed values.
    # Palette does not pre-validate keys/types/values; the API surfaces any errors.
    # override_cluster_api_config = <<-EOT
    #   GCPManagedControlPlane:
    #     spec:
    #       releaseChannel: REGULAR
    # EOT
  }
  # Use update_worker_pools_in_parallel (plural) - the singular update_worker_pool_in_parallel is
  # deprecated and will be removed.
  update_worker_pools_in_parallel = true
  machine_pool {
    name          = "worker-basic"
    count         = 3
    instance_type = "n2-standard-4"

    # Optional: YAML passthrough for pool-level CAPG-managed properties
    # (e.g. GCPManagedMachinePool).
    # override_cluster_api_config = <<-EOT
    #   GCPManagedMachinePool:
    #     spec:
    #       nodePoolName: worker-basic
    # EOT
  }
}
