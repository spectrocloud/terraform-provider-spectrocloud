# Comprehensive spectrocloud_cluster_aks example - exercises the full attribute surface of the
# resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_aks for the minimal happy-path version.
#
# Day-2 mutability summary: `name`, `cloud_account_id`, and every single field inside
# `cloud_config` are ForceNew - AKS's networking and identity configuration cannot be changed
# in place. Everything else - cluster_profile, machine_pool, backup_policy, scan_policy,
# cluster_rbac_binding, namespaces, host_config, and the top-level Day-2 settings - updates in
# place.
#
# Note: unlike most other cluster resources in this provider, machine_pool here has no
# control_plane/control_plane_as_worker fields - AKS's control plane is fully managed by Azure,
# so every machine_pool block is a worker/system node pool.
resource "spectrocloud_cluster_aks" "cluster" {
  # Required, ForceNew.
  name = "e2e-aks-cluster"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster exercising the full AKS schema."
  # Optional. Arbitrary metadata, e.g. for external tooling to key off of.
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  # Required, ForceNew.
  cloud_account_id = spectrocloud_cloudaccount_azure.account.id

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.aks_profile.id
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the aks end-to-end usecase example."
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
  #     id        = spectrocloud_cluster_profile.aks_profile.id
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

  cloud_config {
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure_region
    ssh_key         = var.cluster_ssh_public_key

    # Optional, default false. Creates the AKS API server with a private (non-internet-
    # routable) endpoint.
    private_cluster = false

    # Custom (bring-your-own) VNet - a flat structure (unlike Azure's nested subnet blocks),
    # because AKS static placement doesn't yet support multiple subnets per role on the backend.
    # Omit all of these to let Palette manage networking automatically.
    vnet_name                                = "e2e-aks-vnet"
    vnet_resource_group                      = var.azure_resource_group
    vnet_cidr_block                          = "10.1.0.0/16"
    worker_subnet_name                       = "worker-subnet"
    worker_cidr                              = "10.1.1.0/24"
    worker_subnet_security_group_name        = "worker-nsg"
    control_plane_subnet_name                = "control-plane-subnet"
    control_plane_cidr                       = "10.1.0.0/24"
    control_plane_subnet_security_group_name = "control-plane-nsg"

    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT
  }

  # System node pool - AKS requires at least one. is_system_node_pool, disk_size_gb, and
  # storage_account_type are all Required for every machine_pool.
  machine_pool {
    name                 = "system-pool"
    count                = 3
    instance_type        = "Standard_D4s_v3"
    disk_size_gb         = 80
    is_system_node_pool  = true
    storage_account_type = "Premium_LRS"
    os_sku               = "Ubuntu"
    os_type              = "Linux"
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

    count = 2
    min   = 2
    max   = 6

    instance_type        = "Standard_D2s_v3"
    disk_size_gb         = 100
    is_system_node_pool  = false
    storage_account_type = "StandardSSD_LRS"
    os_sku               = "Ubuntu"
    os_type              = "Linux"

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
      AzureManagedMachinePool:
        spec:
          osDiskSizeGB: 100
    EOT

    # node {
    #   node_id = "worker-pool-node-1"
    #   action  = "cordon"
    # }
  }

  # Windows worker pool - demonstrates os_type = "Windows" / os_sku = "Windows2022".
  machine_pool {
    name                 = "windows-pool"
    count                = 1
    instance_type        = "Standard_D2s_v3"
    disk_size_gb         = 100
    is_system_node_pool  = false
    storage_account_type = "Standard_LRS"
    os_sku               = "Windows2022"
    os_type              = "Windows"
  }

  backup_policy {
    prefix                         = "e2e-aks-backup"
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
