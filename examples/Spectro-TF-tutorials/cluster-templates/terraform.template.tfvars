#####################
# Palette Settings
#####################
palette-project = "Default" # The name of your project in Palette.

##############################
# Hello Universe Configuration
##############################
app_port = 8080 # The cluster port number on which the service will listen for incoming traffic.

create_new_profile_version = false # Set to true to create cluster profile version 1.1.0 with Kubecost.

update_template_profile_version = false # Set to true after profile v1.1.0 is created to update the cluster template.

upgrade_now_timestamp = "" # Set to the current RFC3339 timestamp (for example, "2026-05-12T14:30:00Z") to trigger an immediate upgrade of all clusters launched from the template. Change the value each time you want to re-trigger an upgrade.

###########################
# AWS Deployment Settings
############################
deploy-aws = false # Set to true to deploy to AWS.

aws-cloud-account-name = "REPLACE ME"
aws-region             = "REPLACE ME"
aws-key-pair-name      = "REPLACE ME"

aws_control_plane_nodes = {
  count              = "1"
  control_plane      = true
  instance_type      = "m4.2xlarge"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE ME"]
}

aws_worker_nodes = {
  count              = "1"
  control_plane      = false
  instance_type      = "m4.2xlarge"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE ME"]
}

###########################
# Azure Deployment Settings
############################
deploy-azure = false # Set to true to deploy to Azure.

azure-cloud-account-name = "REPLACE ME"
azure_subscription_id    = "REPLACE ME"
azure_resource_group     = "REPLACE ME"
azure-use-azs            = false
azure-region             = "REPLACE ME"

azure_control_plane_nodes = {
  count               = "1"
  control_plane       = true
  instance_type       = "Standard_D4s_v5"
  disk_size_gb        = "60"
  azs                 = ["REPLACE ME"]
  is_system_node_pool = false
}

azure_worker_nodes = {
  count               = "1"
  control_plane       = false
  instance_type       = "Standard_D4s_v5"
  disk_size_gb        = "60"
  azs                 = ["REPLACE ME"]
  is_system_node_pool = false
}
