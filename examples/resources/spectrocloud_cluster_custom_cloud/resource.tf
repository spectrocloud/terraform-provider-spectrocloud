locals {
  nutanix_cluster_name = "test-tf-nutanix-cluster"
  # Cloud Configurations
  cloud_config_override_variables = {
    CLUSTER_NAME = local.nutanix_cluster_name
    NUTANIX_ADDITIONAL_TRUST_BUNDLE = "test-bundle"
    CONTROL_PLANE_ENDPOINT_IP = "123.12.12.12"
    CONTROL_PLANE_ENDPOINT_PORT = 6443
    NUTANIX_ENDPOINT = "https://test-app.nutanix.com"
    NUTANIX_INSECURE = false
    NUTANIX_PORT = 6443
  }
  # Node Pool config variables
  node_pool_config_variables = {
    MASTER_NODE_POOL_NAME = "master-pool"
    CLUSTER_NAME = local.cloud_config_override_variables["CLUSTER_NAME"]
    CONTROL_PLANE_ENDPOINT_IP = local.cloud_config_override_variables["CONTROL_PLANE_ENDPOINT_IP"]
    NUTANIX_SSH_AUTHORIZED_KEY = "ssh -a test-test"
    KUBERNETES_VERSION = "1.24.0"
    NUTANIX_PRISM_ELEMENT_CLUSTER_NAME = "nutanix-prism"
    NUTANIX_MACHINE_TEMPLATE_IMAGE_NAME = "test-image.iso"
    NUTANIX_SUBNET_NAME = "subnet-test"

    TLS_CIPHER_SUITES ="TLS_256"
    CONTROL_PLANE_ENDPOINT_PORT = local.cloud_config_override_variables["CONTROL_PLANE_ENDPOINT_PORT"]
    KUBEVIP_SVC_ENABLE = false
    KUBEVIP_LB_ENABLE = false
    KUBEVIP_SVC_ELECTION = false
    NUTANIX_MACHINE_BOOT_TYPE = "legacy"
    NUTANIX_MACHINE_MEMORY_SIZE = "4Gi"
    NUTANIX_SYSTEMDISK_SIZE = "40Gi"
    NUTANIX_MACHINE_VCPU_SOCKET = 2
    NUTANIX_MACHINE_VCPU_PER_SOCKET = 1

    WORKER_NODE_POOL_NAME = "worker-pool"
    WORKER_NODE_SIZE = 1
  }
  location = {
    latitude  = 0
    longitude = 0
  }
}

data "spectrocloud_cloudaccount_custom" "nutanix_account" {
  name = "test-tf-demo"
  cloud = "nutanix"
}

data "spectrocloud_cluster_profile" "profile" {
  name = "test-tf-ntix-profile"
  context = "tenant"
}


resource "spectrocloud_cluster_custom_cloud" "cluster_nutanix" {
  name        = local.cloud_config_override_variables.CLUSTER_NAME
  cloud       = "nutanix"
  context     = "tenant"
  tags        = ["dev", "department:tf", "owner:admin"]
  description = "The nutanix cluster with k8 infra profile test"
  cloud_account_id = data.spectrocloud_cloudaccount_custom.nutanix_account.id
  apply_setting = "DownloadAndInstall"
  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_config {
    values = templatefile("config_templates/cloud_config.yaml", local.cloud_config_override_variables)
  }

  machine_pool {
    additional_labels = {
      "owner"   = "tf"
      "purpose" = "testing"
      "type"    = "master"
    }
    control_plane = true
    control_plane_as_worker = true
    node_pool_config = templatefile("config_templates/master_pool_config.yaml", local.node_pool_config_variables)
  }

  machine_pool {
    additional_labels = {
      "owner"   = "tf"
      "purpose" = "testing"
      "type"    = "worker"
    }
    control_plane = false
    control_plane_as_worker = false
    taints {
      key    = "taintkey2"
      value  = "taintvalue2"
      effect = "NoSchedule"
    }
    node_pool_config = templatefile("config_templates/worker_pool_config.yaml", local.node_pool_config_variables)
  }

  cluster_rbac_binding {
    type = "ClusterRoleBinding"

    role = {
      kind = "ClusterRole"
      name = "testRole3"
    }
    subjects {
      type = "User"
      name = "testRoleUser3"
    }
    subjects {
      type = "Group"
      name = "testRoleGroup3"
    }
    subjects {
      type      = "ServiceAccount"
      name      = "testrolesubject3"
      namespace = "testrolenamespace"
    }
  }

  namespaces {
    name = "test5ns"
    resource_allocation = {
      cpu_cores  = "2"
      memory_MiB = "2048"
    }
  }
  // pause_agent_upgrades = "lock"
  os_patch_on_boot = true
  os_patch_schedule = "0 0 * * SUN"
  os_patch_after = "2025-02-14T13:09:21+05:30"
  skip_completion = true
  force_delete = true
  location_config {
    latitude  = local.location["latitude"]
    longitude = local.location["longitude"]
  }
}