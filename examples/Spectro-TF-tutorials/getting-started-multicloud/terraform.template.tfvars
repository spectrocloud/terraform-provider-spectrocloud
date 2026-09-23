#####################
# Palette Settings
#####################
palette-project = "Default" # The name of your project in Palette.

##############################
# Hello Universe Configuration
##############################
app_namespace   = "hello-universe" # The namespace in which the application will be deployed.
app_port        = 8080             # The cluster port number on which the service will listen for incoming traffic.
replicas_number = 1                # The number of pods to be created.
db_password     = "REPLACE ME"     # The database password to connect to the API database.
auth_token      = "REPLACE ME"     # The auth token for the API connection.

###########################
# AWS Deployment Settings
############################
deploy-aws          = false # Set to true to deploy to AWS.
deploy-aws-kubecost = false # Set to true to deploy to AWS and include Kubecost to your cluster profile.

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

###########################
# Azure Deployment Settings
############################
deploy-azure          = false # Set to true to deploy to Azure.
deploy-azure-kubecost = false # Set to true to deploy to Azure and include Kubecost to your cluster profile.
azure-use-azs         = true  # Set to false when you deploy to a region without AZs.

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

###########################
# GCP Deployment Settings
############################
deploy-gcp          = false # Set to true to deploy to GCP.
deploy-gcp-kubecost = false # Set to true to deploy to GCP and include Kubecost to your cluster profile.

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

############################
# VMware Deployment Settings
############################
deploy-vmware          = false # Set to true to deploy to VMware.
deploy-vmware-kubecost = false # Set to true to deploy to VMware and include Kubecost to your cluster profile.

metallb_ip         = "REPLACE ME" # Provide a range of IP addresses for your Metallb load balancer. This range must be included in the PCG's static IP pool range if using static IP placement.
pcg_name           = "REPLACE ME" # Provide the name of the PCG that will be used to deploy the Palette cluster.
datacenter_name    = "REPLACE ME" # Provide the name of the datacenter in vSphere.
folder_name        = "REPLACE ME" # Provide the name of the folder in vSphere.
search_domain      = "REPLACE ME" # Provide the name of the network search domain.
vsphere_cluster    = "REPLACE ME" # Provide the cluster name for the machine pool as it appears in vSphere.
datastore_name     = "REPLACE ME" # Provide the datastore name for the machine pool as it appears in vSphere.
network_name       = "REPLACE ME" # Provide the network name for the machine pool as it appears in vSphere.
resource_pool_name = "REPLACE ME" # Provide the resource pool name for the machine pool as it appears in vSphere.
ssh_key            = ""           # Provide the path to your public SSH key. If not provided, a new key pair will be created.
ssh_key_private    = ""           # Provide the path to your private SSH key. If not provided, a new key pair will be created.

# Static IP Pool Variables - Required for static IP placement only.
deploy-vmware-static = false          # Set to true to deploy to VMware using static IP placement.
network_gateway      = "REPLACE ME"   # Provide the IP address of the vSphere network gateway.
network_prefix       = 0              # Provide the prefix of your vSphere network. Valid values are network CIDR subnet masks from the range 0-32. Example: 18.
ip_range_start       = "REPLACE ME"   # Provide the first IP address of your PCG IP pool range.
ip_range_end         = "REPLACE ME"   # Provide the second IP address of your PCG IP pool range.
nameserver_addr      = ["REPLACE ME"] # Provide a comma-separated list of DNS name server IP addresses.
