# Comprehensive spectrocloud_cluster_brownfield example - exercises the full attribute surface
# of the resource in one registration, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_brownfield for the minimal happy-path version.
#
# Unlike every other resource in this repo, this does not provision any infrastructure - it
# registers an already-running Kubernetes cluster with Palette. There is no cloud_account_id, no
# cloud_config, and no kubeconfig/admin_kube_config output; instead, applying this resource
# produces a `manifest_url` and `kubectl_command` (see outputs.tf) that must be run against the
# real, existing cluster to complete the import.
#
# Day-2 mutability: nothing is marked ForceNew in the schema (Terraform will always try an
# in-place update), but per the provider's own docs, `context` and `import_mode` "cannot be
# updated after creation" - changing them will not trigger a replacement, so treat both as
# effectively set-once. `cluster_profile`, `machine_pool` (node cordon/uncordon), `backup_policy`,
# `scan_policy`, `cluster_rbac_binding`, `namespaces`, `host_config`, and the top-level Day-2
# settings below all update in place.
resource "spectrocloud_cluster_brownfield" "cluster" {
  # Required. Cannot be updated after creation.
  name = "e2e-brownfield-cluster"
  # Optional. Tags in `key:value` form. The `tags` attribute is being phased out in favor of a
  # future `tags_map` - shown here since it's still the supported field today.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase registration exercising the full brownfield schema."

  # Required. Not updatable after creation. Intended values: aws, eks-anywhere, azure, gcp,
  # vsphere, openshift, generic, maas - "generic" covers anything else, including on-prem/DIY
  # Kubernetes. Note: validation for this field is currently disabled in the provider, so any
  # string is technically accepted; use one of the above for a value Palette actually recognizes.
  cloud_type = "generic"

  # Optional, default "project". Allowed: "project", "tenant". Not updatable after creation.
  context = "project"
  # Optional, default "full" (documented; empty string on the wire). Allowed: "read_only", "full".
  # Not updatable after creation.
  import_mode = "full"

  # Optional. Only relevant for clusters behind an outbound HTTP(S) proxy - supported for
  # generic clusters. Not updatable after creation.
  # proxy    = "http://proxy.mycompany.com:3128"
  # no_proxy = "localhost,127.0.0.1,.svc,.cluster.local"

  # Optional. Only needed when the cluster's proxy uses a custom CA certificate that must be
  # mounted into Palette's in-cluster agent. Not updatable after creation.
  # host_path            = "/etc/pki/ca-trust/source/anchors/proxy-ca.crt"
  # container_mount_path = "/etc/ssl/certs/proxy-ca.crt"

  # Optional, default "". Must be an IANA timezone.
  cluster_timezone = "America/New_York"

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.brownfield_profile.id
    # Supplies values for the profile_variables declared on the profile.
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the brownfield end-to-end usecase example."
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

  # Optional. Set to the current time to trigger an immediate control-plane cert renewal on this
  # apply - it is a trigger, not a schedule, so it is left unset here.
  # renew_k8s_certificates_now = timestamp()

  # Optional, default false/20. Force-delete without waiting for cloud resources to clean up
  # first (Palette simply stops managing the cluster; nothing is actually torn out on the
  # brownfield side). force_delete_delay only applies when force_delete is true (minimum 20
  # minutes).
  # force_delete       = true
  # force_delete_delay = 25

  # Optional, default false. Register asynchronously without waiting for import to finish.
  # skip_completion = true

  # Optional. Day-2 node maintenance only - cordon/uncordon nodes that already exist on the
  # imported cluster. Unlike other cluster resources, machine_pool here does not create or size
  # node pools; `name` just groups the node actions below.
  machine_pool {
    name = "worker-pool"
    node {
      node_name = "e2e-worker-node-1"
      action    = "uncordon"
    }
  }

  machine_pool {
    name = "control-plane-pool"
    node {
      node_name = "e2e-control-plane-node-1"
      action    = "uncordon"
    }
  }

  backup_policy {
    prefix             = "e2e-brownfield-backup"
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

  # location_config is Computed-only for this resource - see outputs.tf.
}
