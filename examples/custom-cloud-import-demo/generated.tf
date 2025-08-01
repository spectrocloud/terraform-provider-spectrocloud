# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.

# __generated__ by Terraform from "67d1d6ada3385f6afd736094:project:awstkov2"
resource "spectrocloud_cluster_custom_cloud" "capi_cluster" {
  apply_setting        = "DownloadAndInstall"
  cloud                = "awstkov2"
  cloud_account_id     = "67d18ac2b20f5f385a1c802c"
  context              = "project"
  description          = null
  force_delete         = false
  force_delete_delay   = 20
  name                 = "poc-k8s-internal-e1"
  os_patch_after       = null
  os_patch_on_boot     = false
  os_patch_schedule    = null
  pause_agent_upgrades = "unlock"
  skip_completion      = false
  tags                 = ["Applications:k8s"]
  cloud_config {
    overrides = {}
    values    = file("cluster_configs_yaml/capi_cluster_cloud_config.yaml")
  }
  cluster_profile {
    id        = "67d073a1a3385f1401455d88"
    variables = {}
  }
  cluster_profile {
    id        = "67d47fc0b20f5ff72c60b9ad"
    variables = {}
  }
  cluster_profile {
    id        = "67e1c463b20f6408831a88eb"
    variables = {}
  }
  location_config {
    country_code = null
    country_name = null
    latitude     = 0
    longitude    = 0
    region_code  = null
    region_name  = null
  }
  machine_pool {
    control_plane           = true
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-enc-control-plane"
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 100
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-agent-md-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-agent-enc-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-agent-md-2a"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-agent-md-2a"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-agent-md-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-agent-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-agent-md-2b"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-agent-md-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-agent-md-2c_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-agent-enc-2c"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-agent-md-2c"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-agent-md-2c"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-ngx-int-md-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-ngx-int-enc-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-ngx-int-md-2a"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-ngx-int-md-2a"
      REPLICAS              = 2
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "nginx-internal"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-ngx-int-md-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-ngx-int-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-ngx-int-md-2b"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-ngx-int-md-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "nginx-internal"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-ngx-int-md-2c_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-ngx-int-enc-2c"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-ngx-int-md-2c"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-ngx-int-md-2c"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "nginx-internal"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-thanos-md-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-thanos-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-thanos-md-2b"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-thanos-md-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "r5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "thanos-store"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-prometheus-md-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-prometheus-enc-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-prometheus-md-2a"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-prometheus-md-2a"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "r5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "prometheus"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-prometheus-md-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-prometheus-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-prometheus-md-2b"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-prometheus-md-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "r5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "prometheus"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-istio-md-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-istio-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-istio-md-2b"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-istio-md-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-ngx-ext-md-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-ngx-ext-enc-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-ngx-ext-md-2a"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-ngx-ext-md-2a"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "nginx-external"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-ngx-ext-md-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-ngx-ext-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-ngx-ext-md-2b"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-ngx-ext-md-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "nginx-external"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-ngx-ext-md-2c_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-ngx-ext-enc-2c"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-ngx-ext-md-2c"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-ngx-ext-md-2c"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
    taints {
      effect = "NoSchedule"
      key    = "nodetaint"
      value  = "nginx-external"
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-istio-ext-md-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-istio-ext-enc-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-istio-ext-md-2a"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-istio-ext-md-2a"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-storage-enc-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-storage-enc-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-storage-enc-2a"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-storage-enc-2b_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-storage-enc-2b"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-storage-enc-2b"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-storage-enc-2c_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-storage-enc-2c"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-storage-enc-2c"
      REPLICAS              = 1
      AMI_ID                = "ami-04d17236f2120c048"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
  machine_pool {
    control_plane           = false
    control_plane_as_worker = false
    node_pool_config        = file("cluster_configs_yaml/capi_cluster_poc-k8s-internal-e1-agent-test-2a_config.yaml")
    overrides = {
      MACHINE_TEMPLATE_NAME = "poc-k8s-internal-e1-agent-test-5-2a"
      KC_TEMPLATE_NAME      = "poc-k8s-internal-e1-agent-test-2a"
      NODE_POOL_NAME        = "poc-k8s-internal-e1-agent-test-2a"
      REPLICAS              = 1
      AMI_ID                = "ami-00d9648e3efaffa25"
      INSTANCE_TYPE         = "m5.4xlarge"
      ROOT_VOLUME_SIZE      = 60
    }
  }
}
