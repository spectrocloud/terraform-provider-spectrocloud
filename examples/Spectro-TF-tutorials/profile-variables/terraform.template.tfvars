#####################
# Palette Settings
#####################
palette-project = "Default" # The name of your project in Palette.

##############################
# Application Configuration
##############################
wordpress_replica   = "REPLACE ME" # The number of pods to be created for Wordpress.
wordpress_namespace = "REPLACE ME" # The namespace to be created for Wordpress.
wordpress_port      = "REPLACE ME" # The port to be created for HTTP for Wordpress.

############################
# AWS Deployment Settings
############################
deploy-aws     = false # Set to true to deploy to AWS.
deploy-aws-var = false # Set to true to deploy to AWS with cluster profile variables.

aws-cloud-account-name = "REPLACE ME"
aws-region             = "REPLACE ME"
aws-key-pair-name      = "REPLACE ME"

aws_control_plane_nodes = {
  count              = "1"
  control_plane      = true
  instance_type      = "m4.2xlarge"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-east-1a", "us-east-1b"].
}

aws_worker_nodes = {
  count              = "1"
  control_plane      = false
  instance_type      = "m4.2xlarge"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-east-1a", "us-east-1b"].
}

############################
# Azure Deployment Settings
############################
deploy-azure     = false # Set to true to deploy to Azure.
deploy-azure-var = false # Set to true to deploy to Azure with cluster profile variables
azure-use-azs    = false # Set to false when you deploy to a region without AZs.

azure-cloud-account-name = "REPLACE ME"
azure-region             = "REPLACE ME"
azure_subscription_id    = "REPLACE ME"
azure_resource_group     = "REPLACE ME"


azure_control_plane_nodes = {
  count               = "1"
  control_plane       = true
  instance_type       = "Standard_D4s_v3"
  disk_size_gb        = "60"
  azs                 = ["1"] # If you want to deploy to multiple AZs, add them here.
  is_system_node_pool = false
}

azure_worker_nodes = {
  count               = "1"
  control_plane       = false
  instance_type       = "Standard_D4s_v3"
  disk_size_gb        = "60"
  azs                 = ["1"] # If you want to deploy to multiple AZs, add them here.
  is_system_node_pool = false
}

############################
# GCP Deployment Settings
############################
deploy-gcp     = false # Set to true to deploy to GCP.
deploy-gcp-var = false # Set to true to deploy to GCP with cluster profile variables.

gcp-cloud-account-name = "REPLACE ME"
gcp-region             = "REPLACE ME"
gcp_project_name       = "REPLACE ME"

gcp_control_plane_nodes = {
  count              = "1"
  control_plane      = true
  instance_type      = "n2-standard-4"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-central1-a", "us-central1-b"].
}

gcp_worker_nodes = {
  count              = "1"
  control_plane      = false
  instance_type      = "n2-standard-4"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-central1-a", "us-central1-b"].
}
