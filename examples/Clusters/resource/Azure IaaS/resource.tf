data "spectrocloud_cluster_profile" "profile" {
  name = "tf-js13-azure-profile"
}

data "spectrocloud_cloudaccount_azure" "account" {
  name = "jayesh-azure-ca"
}

# Day-2 mutability: only `name` and `cloud_account_id` (below) are ForceNew - changing either
# recreates the cluster. cluster_profile, machine_pool, backup_policy, and scan_policy update in
# place; see cloud_config's own header below for its mutability.
resource "spectrocloud_cluster_azure" "cluster" {
  name = "tf-azure-js-1"
  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }
  cloud_account_id = data.spectrocloud_cloudaccount_azure.account.id

  # cloud_config:
  #   subscription_id, resource_group, region, ssh_key - NOT ForceNew (unlike AWS/AKS); these
  #     update in place along with everything else in this block.
  #   override_cluster_api_config - Optional. Raw Cluster API config overrides, merged into the
  #     generated cluster spec.
  #   network_resource_group/virtual_network_name/virtual_network_cidr_block - Optional,
  #     alternative to Palette-managed networking: bring-your-own VNet. Required together
  #     (RequiredWith) if control_plane_subnet, worker_node_subnet, or private_api_server is set.
  #   control_plane_subnet / worker_node_subnet - Optional nested blocks for bring-your-own
  #     subnet placement; each takes name/cidr_block, with security_group_name optional.
  #   private_api_server - Optional. Gives the cluster a private API server endpoint instead of a
  #     public one; resource_group is required, private_dns_zone and static_ip are optional
  #     (auto-created/allocated if omitted).
  cloud_config {
    subscription_id             = var.azure_subscription_id
    resource_group              = var.azure_resource_group
    region                      = var.azure_region
    ssh_key                     = var.cluster_ssh_public_key
    override_cluster_api_config = <<-EOT
      spec:
        controlPlaneConfiguration:
          apiServer:
            extraArgs:
              authorization-mode: Node,RBAC
    EOT

    # network_resource_group     = "test-resource-group"
    # virtual_network_name       = "test-network-name"
    # virtual_network_cidr_block = "10.0.0.9/10"
    # control_plane_subnet {
    #   name                = "cp_subnet_name"
    #   cidr_block          = "10.0.0.9/16"
    #   security_group_name = "cp_subnet_security_name"
    # }
    # worker_node_subnet {
    #   name                = "worker_subnet_name"
    #   cidr_block          = "10.0.0.9/16"
    #   security_group_name = "worker_subnet_security_name"
    # }
    # private_api_server {
    #   resource_group   = "test-resource-group"
    #   private_dns_zone = "test-private-dns-zone"
    #   static_ip        = "10.11.12.51"
    # }

  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "cp-pool"
    count                   = 1
    instance_type           = "Standard_D2_v3"
    azs                     = []
    disk {
      size_gb = 65
      type    = "Standard_LRS"
    }
  }

  # machine_pool (worker pool "worker-basic"):
  #   override_cluster_api_config - Optional. Raw Cluster API config overrides scoped to this
  #     pool, merged into the generated cluster spec.
  #   override_health_check_configuration - Optional. Raw Machine Health Check overrides for this
  #     node pool.
  machine_pool {
    is_system_node_pool         = true
    name                        = "worker-basic"
    count                       = 2
    instance_type               = "Standard_D2_v3"
    azs                         = []
    override_cluster_api_config = <<-EOT
      spec:
        template:
          spec:
            nodeDrainTimeout: 5m
    EOT

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
}
