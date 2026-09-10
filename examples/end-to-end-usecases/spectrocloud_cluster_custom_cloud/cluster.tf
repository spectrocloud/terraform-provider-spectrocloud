locals {
  custom_cloud_cluster_name = "e2e-custom-cloud-cluster"

  # Values for config_templates/cloud_config.yaml - one Kubernetes Secret/ConfigMap/NutanixCluster
  # manifest set describing how Palette should reach this Nutanix Prism Central.
  cloud_config_template_vars = {
    CLUSTER_NAME                    = local.custom_cloud_cluster_name
    NUTANIX_ADDITIONAL_TRUST_BUNDLE = "e2e-demo-trust-bundle"
    CONTROL_PLANE_ENDPOINT_IP       = var.control_plane_endpoint_ip
    CONTROL_PLANE_ENDPOINT_PORT     = 6443
    NUTANIX_ENDPOINT                = var.nutanix_endpoint
    NUTANIX_INSECURE                = false
    NUTANIX_PORT                    = var.nutanix_port
  }

  # Values shared by config_templates/cp_pool_config.yaml and worker_pool_config.yaml.
  node_pool_template_vars = {
    CP_NODE_POOL_NAME                   = "cp-pool"
    CLUSTER_NAME                        = local.custom_cloud_cluster_name
    CONTROL_PLANE_ENDPOINT_IP           = var.control_plane_endpoint_ip
    CONTROL_PLANE_ENDPOINT_PORT         = local.cloud_config_template_vars["CONTROL_PLANE_ENDPOINT_PORT"]
    NUTANIX_SSH_AUTHORIZED_KEY          = var.nutanix_ssh_authorized_key
    KUBERNETES_VERSION                  = "1.28.5"
    NUTANIX_PRISM_ELEMENT_CLUSTER_NAME  = var.nutanix_prism_element_cluster_name
    NUTANIX_MACHINE_TEMPLATE_IMAGE_NAME = var.nutanix_machine_template_image_name
    NUTANIX_SUBNET_NAME                 = var.nutanix_subnet_name
    TLS_CIPHER_SUITES                   = "TLS_AES_256_GCM_SHA384"
    KUBEVIP_SVC_ENABLE                  = false
    KUBEVIP_LB_ENABLE                   = false
    KUBEVIP_SVC_ELECTION                = false
    NUTANIX_MACHINE_BOOT_TYPE           = "legacy"
    NUTANIX_MACHINE_MEMORY_SIZE         = "4Gi"
    NUTANIX_SYSTEMDISK_SIZE             = "40Gi"
    NUTANIX_MACHINE_VCPU_SOCKET         = 2
    NUTANIX_MACHINE_VCPU_PER_SOCKET     = 1
    WORKER_NODE_POOL_NAME               = "worker-pool"
    WORKER_NODE_SIZE                    = 2
  }
}

# Comprehensive spectrocloud_cluster_custom_cloud example - exercises the full attribute surface
# of the resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_custom_cloud for the minimal happy-path version.
#
# Unlike every other cluster resource in this repo, a custom cloud cluster has no structured
# cloud_config/machine_pool schema at all - node provisioning is described as raw Cluster API
# YAML (config_templates/*.yaml, rendered with templatefile()), with an optional `overrides` map
# on both cloud_config and machine_pool for patching that YAML without re-templating it (template
# variables, wildcard field matches, or Kind.path/global path syntax).
#
# Day-2 mutability: `name`, `cloud`, and `cloud_account_id` are ForceNew - changing any of them
# recreates the cluster. `cloud_config.values`/`overrides` and `machine_pool.node_pool_config`/
# `overrides` are NOT ForceNew and update in place, along with cluster_profile, backup_policy,
# scan_policy, namespaces, cluster_rbac_binding, and the top-level Day-2 settings below.
resource "spectrocloud_cluster_custom_cloud" "cluster" {
  # Required, ForceNew.
  name = local.custom_cloud_cluster_name
  # Required, ForceNew. The custom cloud provider name - must match the cloud account's `cloud`.
  cloud = "nutanix"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster exercising the full custom cloud (Nutanix) schema."

  # Required, ForceNew. References the cloud account created in cloudaccount.tf.
  cloud_account_id = spectrocloud_cloudaccount_custom.account.id

  # Optional, Computed. Can only be set at creation - "PureManage" (Palette-provisioned,
  # default) vs "PureAttach" (attach an already-running cluster).
  cluster_type = "PureManage"

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.custom_cloud_profile.id
    # Supplies values for the profile_variables declared on the profile.
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the custom-cloud end-to-end usecase example."
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
  #     id        = spectrocloud_cluster_profile.custom_cloud_profile.id
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

  # Required. The cluster-scoped Cluster API manifests (Secret, ConfigMap, NutanixCluster,
  # MachineHealthCheck), fully rendered via templatefile() from config_templates/cloud_config.yaml.
  cloud_config {
    values = templatefile("${path.module}/config_templates/cloud_config.yaml", local.cloud_config_template_vars)

    # Optional: patch a value in the rendered YAML above without re-templating the whole file.
    # Field pattern search updates every field named "insecure" anywhere in the YAML document.
    overrides = {
      "insecure" = "false"
    }
  }

  # Control-plane pool: the manifests (KubeadmControlPlane, NutanixMachineTemplate) come from
  # config_templates/cp_pool_config.yaml.
  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    node_pool_config        = templatefile("${path.module}/config_templates/cp_pool_config.yaml", local.node_pool_template_vars)

    # Optional: patch the rendered YAML without re-templating it. Document-specific syntax
    # ("Kind.path") targets a field only within matching document kinds.
    overrides = {
      "KubeadmControlPlane.spec.replicas" = "1"
    }

    taints {
      key    = "dedicated"
      value  = "control-plane"
      effect = "NoSchedule"
    }
  }

  # Worker pool: the manifests (MachineDeployment, NutanixMachineTemplate, KubeadmConfigTemplate)
  # come from config_templates/worker_pool_config.yaml.
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = templatefile("${path.module}/config_templates/worker_pool_config.yaml", local.node_pool_template_vars)

    # Global path syntax (no "Kind." prefix) updates the first matching field across all
    # documents in the rendered YAML.
    overrides = {
      "spec.replicas" = "2"
    }

    taints {
      key    = "dedicated"
      value  = "general"
      effect = "NoSchedule"
    }
  }

  backup_policy {
    prefix             = "e2e-custom-cloud-backup"
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

  # Unlike most other cluster resources, location_config is settable here (not Computed-only) -
  # Palette can't otherwise infer a custom cloud's physical location from raw CAPI YAML.
  location_config {
    latitude  = 37.3861
    longitude = -122.0839
    # country_code = "US"
  }
}
