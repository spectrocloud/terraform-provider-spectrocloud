###########################
# AWS Deployment Settings
############################
deploy-aws = false # Set to true to deploy to AWS

aws-cloud-account-name = "REPLACE_ME"
aws-region             = "REPLACE_ME"
aws-key-pair-name      = "REPLACE_ME"

aws-hello-universe-api-uri = "http://REPLACE_ME:3000" # Set IP address of hello-universe API once deployed

aws_control_plane_nodes = {
  count              = "1"
  control_plane      = true
  instance_type      = "m4.2xlarge"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE_ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-east-1a", "us-east-1b"]
}

aws_worker_nodes = {
  count              = "1"
  control_plane      = false
  instance_type      = "m4.2xlarge"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE_ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-east-1a", "us-east-1b"]
}

###########################
# Azure Deployment Settings
############################
deploy-azure  = false # Set to true to deploy to Azure
azure-use-azs = true  # Set to false when you deploy to a region without AZs

azure-cloud-account-name = "REPLACE_ME"
azure-region             = "REPLACE_ME"
azure_subscription_id    = "REPLACE_ME"
azure_resource_group     = "REPLACE_ME"

azure-hello-universe-api-uri = "http://REPLACE_ME:3000" # Set IP address of hello-universe API once deployed

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

###########################
# GCP Deployment Settings
############################
deploy-gcp = false # Set to true to deploy to GCP

gcp-cloud-account-name = "REPLACE_ME"
gcp-region             = "REPLACE_ME"
gcp_project_name       = "REPLACE_ME"

gcp-hello-universe-api-uri = "http://REPLACE_ME:3000" # Set IP address of hello-universe API once deployed

gcp_control_plane_nodes = {
  count              = "1"
  control_plane      = true
  instance_type      = "n2-standard-4"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE_ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-central1-a", "us-central1-b"]
}

gcp_worker_nodes = {
  count              = "1"
  control_plane      = false
  instance_type      = "n2-standard-4"
  disk_size_gb       = "60"
  availability_zones = ["REPLACE_ME"] # If you want to deploy to multiple AZs, add them here. Example: ["us-central1-a", "us-central1-b"]
}

###########################
# Day-2 Profile Update
###########################
update_profile_version = false # Set to true after retrieving the API IP(s) above to switch the frontend cluster(s) to the 3-tier profile.
