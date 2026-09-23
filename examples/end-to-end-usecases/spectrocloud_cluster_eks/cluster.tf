# Comprehensive spectrocloud_cluster_eks example - exercises the full attribute surface of the
# resource in one cluster, not just the minimal required fields. See
# examples/resources/spectrocloud_cluster_eks for the minimal happy-path version.
#
# Day-2 mutability summary: `name`/`cloud_account_id` and most of `cloud_config` are ForceNew -
# except `public_access_cidrs`/`private_access_cidrs`, which update in place. Everything else -
# cluster_profile, machine_pool, fargate_profile, backup_policy, scan_policy,
# cluster_rbac_binding, namespaces, host_config, and the top-level Day-2 settings - updates in
# place.
#
# Note: like AKS, EKS's control plane is fully AWS-managed, so machine_pool has no
# control_plane/control_plane_as_worker fields - every pool here is a worker pool (an EKS
# Managed Node Group).
resource "spectrocloud_cluster_eks" "cluster" {
  # Required, ForceNew.
  name    = "e2e-eks-cluster"
  context = "project"
  # Optional, mutually exclusive with the deprecated `tags` - use tags_map.
  tags_map = {
    "e2e-usecase" = "true"
    "team"        = "platform"
  }
  description            = "Comprehensive end-to-end usecase cluster exercising the full EKS schema."
  cluster_meta_attribute = jsonencode({ nic_name = "eth0", env = "e2e-demo" })

  cloud_account_id = spectrocloud_cloudaccount_aws.account.id

  apply_setting = "DownloadAndInstall"

  cluster_profile {
    id = spectrocloud_cluster_profile.eks_profile.id
    variables = {
      app_version      = "1.2.0"
      environment_tier = "staging"
      notes            = "Deployed by the eks end-to-end usecase example."
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
  #     id        = spectrocloud_cluster_profile.eks_profile.id
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
    # Optional, ForceNew.
    ssh_key_name = var.cluster_ssh_key_name
    # Required, ForceNew.
    region = var.aws_region
    # Optional, ForceNew. If unset, Palette provisions a new VPC dynamically.
    # vpc_id = "vpc-0a38a86f3bf3c6cf5"

    # Dynamic AZ selection, ForceNew - mutually exclusive with az_subnets (static, also
    # ForceNew).
    azs = ["us-east-1a", "us-east-1b", "us-east-1c"]
    # az_subnets = {
    #   "us-east-1a" = "subnet-0123456789abcdef0"
    #   "us-east-1b" = "subnet-0123456789abcdef1"
    # }

    # Optional, default "public", ForceNew. Allowed: "public", "private", "private_and_public".
    endpoint_access = "public"
    # Optional. Unlike the rest of cloud_config, these two update in place.
    public_access_cidrs  = ["0.0.0.0/0"]
    private_access_cidrs = ["10.0.0.0/8"]

    # Optional, ForceNew. ARN of a KMS key for EKS secrets encryption.
    # encryption_config_arn = var.eks_kms_key_arn

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
    instance_type = "t3.xlarge"
    disk_size_gb  = 80

    # Autoscaling: when min/max are both set, `count` must equal `min` - the provider rejects a
    # count greater than min. The autoscaler then adjusts the live node count up to `max`.
    count = 2
    min   = 2
    max   = 6

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

    # Optional, default "AL2023_x86_64_STANDARD". AL2_x86_64/AL2_x86_64_GPU are deprecated.
    ami_type = "AL2023_x86_64_STANDARD"

    # Optional, default "on-demand". "spot" uses EC2 spot capacity at max_price.
    capacity_type = "spot"
    max_price     = "0.05"

    azs = ["us-east-1a", "us-east-1b"]
    # az_subnets = {
    #   "us-east-1a" = "subnet-0123456789abcdef0"
    #   "us-east-1b" = "subnet-0123456789abcdef1"
    # }

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
      AWSManagedMachinePool:
        spec:
          diskSize: 100
    EOT

    # Optional: custom EC2 launch template settings for this pool.
    eks_launch_template {
      # Optional. If unset, Palette repaves the cluster automatically on EKS AMI updates.
      # ami_id                 = "ami-0123456789abcdef0"
      root_volume_type           = "gp3"
      root_volume_iops           = 3000
      root_volume_throughput     = 125
      additional_security_groups = ["sg-0123456789abcdef0"]
    }

    # node {
    #   node_id = "worker-pool-node-1"
    #   action  = "cordon"
    # }
  }

  # Optional: Fargate profiles run pods matching the given namespace/label selectors on
  # serverless AWS Fargate instead of the EC2-backed machine pools above.
  fargate_profile {
    name    = "e2e-fargate-profile"
    subnets = var.eks_fargate_subnet_ids
    additional_tags = {
      "team" = "platform"
    }
    selector {
      namespace = "fargate-workloads"
      labels = {
        "compute-type" = "fargate"
      }
    }
  }

  backup_policy {
    prefix                         = "e2e-eks-backup"
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
