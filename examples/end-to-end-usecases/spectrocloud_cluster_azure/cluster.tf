# Comprehensive spectrocloud_cluster_azure example - exercises the full attribute surface of the
# resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_azure for the minimal happy-path version.
#
# Day-2 mutability summary: `name` and `cloud_account_id` are ForceNew, along with every
# cloud_config field individually. Everything else - cluster_profile, machine_pool,
# backup_policy, scan_policy, cluster_rbac_binding, namespaces, host_config, and the top-level
# Day-2 settings - updates in place.
resource "spectrocloud_cluster_azure" "cluster" {
  # Required, ForceNew.
  name = "e2e-azure-cluster"
  # Optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Optional. Tags in `key:value` form.
  tags = ["e2e-usecase", "team:platform", "owner:bob"]
  # Optional, default "". Free-text description.
  description = "Comprehensive end-to-end usecase cluster exercising the full Azure schema."
  # Optional. Arbitrary metadata, e.g. for external tooling to key off of.
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  # Required, ForceNew.
  cloud_account_id = spectrocloud_cloudaccount_azure.account.id

  # Optional, default "DownloadAndInstall". "DownloadAndInstallLater" only downloads artifacts
  # and postpones pack installation.
  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.azure_profile.id
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the azure end-to-end usecase example."
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
  #     id        = spectrocloud_cluster_profile.azure_profile.id
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
    # Required, ForceNew.
    subscription_id = var.azure_subscription_id
    resource_group  = var.azure_resource_group
    region          = var.azure_region
    ssh_key         = var.cluster_ssh_public_key

    # Optional, ForceNew. Storage account/container for cluster diagnostics data.
    storage_account_name = "e2eazurestorage"
    container_name       = "diagnostics"

    # Custom (bring-your-own) VNet - all 4 of these fields are RequiredWith each other; if any
    # one is set, all must be. Omit all 4 to let Palette manage networking automatically.
    network_resource_group     = var.azure_resource_group
    virtual_network_name       = "e2e-azure-vnet"
    virtual_network_cidr_block = "10.0.0.0/16"
    control_plane_subnet {
      name                = "control-plane-subnet"
      cidr_block          = "10.0.0.0/24"
      security_group_name = "control-plane-nsg"
    }
    worker_node_subnet {
      name                = "worker-subnet"
      cidr_block          = "10.0.1.0/24"
      security_group_name = "worker-nsg"
    }

    # Optional: private API server with a custom DNS zone - only valid alongside the custom
    # VNet fields above.
    # private_api_server {
    #   resource_group   = var.azure_resource_group
    #   private_dns_zone = "e2e-demo.privatelink.eastus.azmk8s.io"
    #   static_ip        = "10.0.0.100"
    # }

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
    instance_type           = "Standard_D4s_v3"
    disk {
      size_gb = 80
      type    = "Premium_LRS"
    }
    os_type             = "Linux"
    is_system_node_pool = true
    azs                 = ["1", "2", "3"]
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

    # Note: unlike some other cluster resources, spectrocloud_cluster_azure's machine_pool has
    # no min/max autoscaling fields - count is the pool's fixed size.
    count = 2

    instance_type = "Standard_D2s_v3"
    disk {
      size_gb = 100
      type    = "StandardSSD_LRS"
    }
    os_type             = "Linux"
    is_system_node_pool = false
    azs                 = ["1", "2"]

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
      AzureMachineTemplate:
        spec:
          template:
            spec:
              osDisk:
                diskSizeGB: 100
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

  # Windows worker pool - demonstrates os_type = "Windows".
  machine_pool {
    name          = "windows-pool"
    count         = 1
    instance_type = "Standard_D2s_v3"
    disk {
      size_gb = 100
      type    = "Standard_LRS"
    }
    os_type = "Windows"
    azs     = ["1"]
  }

  backup_policy {
    prefix                         = "e2e-azure-backup"
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
