#########
# Palette
#########

variable "palette-project" {
  type        = string
  description = "The name of your project in Palette."

  validation {
    condition     = var.palette-project != ""
    error_message = "Provide the correct Palette project."
  }

}

#######
# AWS
#######
variable "aws-cloud-account-name" {
  type        = string
  description = "The name of your AWS account as assigned in Palette."

  validation {
    condition     = var.deploy-aws ? var.aws-cloud-account-name != "REPLACE ME" && var.aws-cloud-account-name != "" : true
    error_message = "Provide the correct AWS cloud account name."
  }
}

variable "deploy-aws" {
  type        = bool
  description = "A flag for enabling a deployment on AWS."
}

variable "deploy-aws-kubecost" {
  type        = bool
  description = "A flag for enabling a deployment on AWS with Kubecost."
}

variable "aws-region" {
  type        = string
  description = "AWS region"
  default     = "us-east-1"

  validation {
    condition     = var.deploy-aws ? var.aws-region != "REPLACE ME" && var.aws-region != "" : true
    error_message = "Provide the correct AWS region."
  }
}

variable "aws-key-pair-name" {
  type        = string
  description = "The name of the AWS key pair to use for SSH access to the cluster."

  validation {
    condition     = var.deploy-aws ? var.aws-key-pair-name != "REPLACE ME" && var.aws-key-pair-name != "" : true
    error_message = "Provide the correct AWS SSH key pair name."
  }
}

variable "aws_control_plane_nodes" {
  type = object({
    count              = string
    control_plane      = bool
    instance_type      = string
    disk_size_gb       = string
    availability_zones = list(string)
  })
  default = {
    count              = "1"
    control_plane      = true
    instance_type      = "m4.2xlarge"
    disk_size_gb       = "60"
    availability_zones = ["us-east-1a"]
  }
  description = "AWS control plane nodes configuration."

  validation {
    condition     = var.deploy-aws ? length(var.aws_control_plane_nodes.availability_zones) > 0 && !contains(var.aws_control_plane_nodes.availability_zones, "REPLACE ME") : true
    error_message = "The availability_zones parameter must be set correctly"
  }
}
variable "aws_worker_nodes" {
  type = object({
    count              = string
    control_plane      = bool
    instance_type      = string
    disk_size_gb       = string
    availability_zones = list(string)
  })
  default = {
    count              = "1"
    control_plane      = false
    instance_type      = "m4.2xlarge"
    disk_size_gb       = "60"
    availability_zones = ["us-east-1a"]
  }
  description = "AWS worker nodes configuration."

  validation {
    condition     = var.deploy-aws ? length(var.aws_worker_nodes.availability_zones) > 0 && !contains(var.aws_worker_nodes.availability_zones, "REPLACE ME") : true
    error_message = "The availability_zones parameter must be set correctly"
  }
}

#######
# Azure
#######
variable "azure-cloud-account-name" {
  type        = string
  description = "The name of your Azure account as assigned in Palette."
  default     = ""

  validation {
    condition     = var.deploy-azure ? var.azure-cloud-account-name != "REPLACE ME" && var.azure-cloud-account-name != "" : true
    error_message = "Provide the correct Azure cloud account name."
  }
}

variable "deploy-azure" {
  type        = bool
  description = "A flag for enabling a deployment on Azure."
}

variable "deploy-azure-kubecost" {
  type        = bool
  description = "A flag for enabling a deployment on Azure with Kubecost."
}

variable "azure_subscription_id" {
  type        = string
  description = "Azure subscription ID."
  default     = ""

  validation {
    condition     = var.deploy-azure ? var.azure_subscription_id != "REPLACE ME" && var.azure_subscription_id != "" : true
    error_message = "Provide the correct Azure subscription ID."
  }
}

variable "azure_resource_group" {
  type        = string
  description = "Azure resource group."
  default     = ""

  validation {
    condition     = var.deploy-azure ? var.azure_resource_group != "REPLACE ME" && var.azure_resource_group != "" : true
    error_message = "Provide the correct Azure resource group name."
  }
}

variable "azure-use-azs" {
  type        = bool
  description = "A flag for configuring whether to use Azure Availability Zones. Check if your Azure region supports availability zones by reviewing the [Azure Regions and Availability Zones](https://learn.microsoft.com/en-us/azure/reliability/availability-zones-service-support#azure-regions-with-availability-zone-support) resource."
}

variable "azure-region" {
  type        = string
  description = "Azure region."
  default     = "eastus"

  validation {
    condition     = var.deploy-azure ? var.azure-region != "REPLACE ME" && var.azure-region != "" : true
    error_message = "Provide the correct Azure region name."
  }
}

variable "azure_control_plane_nodes" {
  type = object({
    count               = string
    control_plane       = bool
    instance_type       = string
    disk_size_gb        = string
    azs                 = list(string)
    is_system_node_pool = bool
  })
  default = {
    count               = "1"
    control_plane       = true
    instance_type       = "Standard_A8_v2"
    disk_size_gb        = "60"
    azs                 = ["1"]
    is_system_node_pool = false
  }
  description = "Azure control plane nodes configuration."
}

variable "azure_worker_nodes" {
  type = object({
    count               = string
    control_plane       = bool
    instance_type       = string
    disk_size_gb        = string
    azs                 = list(string)
    is_system_node_pool = bool
  })
  default = {
    count               = "1"
    control_plane       = false
    instance_type       = "Standard_A8_v2"
    disk_size_gb        = "60"
    azs                 = ["1"]
    is_system_node_pool = false
  }
  description = "Azure worker nodes configuration."
}

#######
# GCP
#######
variable "gcp-cloud-account-name" {
  type        = string
  description = "The name of your GCP account as assigned in Palette."
  default     = ""

  validation {
    condition     = var.deploy-gcp ? var.gcp-cloud-account-name != "REPLACE ME" && var.gcp-cloud-account-name != "" : true
    error_message = "Provide the correct GCP cloud account name."
  }
}

variable "gcp_project_name" {
  type        = string
  description = "The name of your GCP project."
  default     = ""

  validation {
    condition     = var.deploy-gcp ? var.gcp_project_name != "REPLACE ME" && var.gcp_project_name != "" : true
    error_message = "Provide the correct GCP project name."
  }
}

variable "deploy-gcp" {
  type        = bool
  description = "A flag for enabling a deployment on GCP."
}

variable "deploy-gcp-kubecost" {
  type        = bool
  description = "A flag for enabling a deployment on GCP with Kubecost."
}

variable "gcp-region" {
  type        = string
  description = "GCP region"
  default     = "us-central1"

  validation {
    condition     = var.deploy-gcp ? var.gcp-region != "REPLACE ME" && var.gcp-region != "" : true
    error_message = "Provide the correct GCP region."
  }
}

variable "gcp_control_plane_nodes" {
  type = object({
    count              = string
    control_plane      = bool
    instance_type      = string
    disk_size_gb       = string
    availability_zones = list(string)
  })
  default = {
    count              = "1"
    control_plane      = true
    instance_type      = "n1-standard-4"
    disk_size_gb       = "60"
    availability_zones = ["us-central1-a"]
  }
  description = "GCP control plane nodes configuration."

  validation {
    condition     = var.deploy-gcp ? length(var.gcp_control_plane_nodes.availability_zones) > 0 && !contains(var.gcp_control_plane_nodes.availability_zones, "REPLACE ME") : true
    error_message = "The availability_zones parameter must be set correctly"
  }
}

variable "gcp_worker_nodes" {
  type = object({
    count              = string
    control_plane      = bool
    instance_type      = string
    disk_size_gb       = string
    availability_zones = list(string)
  })
  default = {
    count              = "1"
    control_plane      = false
    instance_type      = "n1-standard-4"
    disk_size_gb       = "60"
    availability_zones = ["us-central1-a"]
  }
  description = "GCP worker nodes configuration."

  validation {
    condition     = var.deploy-gcp ? length(var.gcp_worker_nodes.availability_zones) > 0 && !contains(var.gcp_worker_nodes.availability_zones, "REPLACE ME") : true
    error_message = "The availability_zones parameter must be set correctly"
  }
}

variable "tags" {
  type        = list(string)
  description = "The default tags to apply to Palette resources."
  default = [
    "spectro-cloud-education",
    "app:hello-universe",
    "spectrocloud:tutorials",
    "terraform_managed:true",
    "tutorial:getting-started-terraform"
  ]
}


########
# VMware
########

variable "deploy-vmware" {
  type        = bool
  description = "A flag for enabling a deployment on VMware."
}

variable "deploy-vmware-kubecost" {
  type        = bool
  description = "A flag for enabling a deployment on VMware with Kubecost."
}

variable "metallb_ip" {
  type        = string
  description = "The IP address range for your MetalLB load balancer."

  validation {
    condition     = var.deploy-vmware ? var.metallb_ip != "REPLACE ME" && var.metallb_ip != "" : true
    error_message = "Provide the correct MetalLB IP."
  }
}

variable "ssh_key" {
  type        = string
  description = "The path to the public key that will be added to the cluster nodes. If not provided, a new key pair will be generated."

  validation {
    condition     = var.ssh_key == "" ? true : fileexists(var.ssh_key)
    error_message = "The provided SSH key file does not exist. Please, provide a valid path."
  }
}

variable "ssh_key_private" {
  type        = string
  description = "The path to the private key that will be used to access the cluster nodes. If not provided, a new key pair will be generated."

  validation {
    condition     = var.ssh_key_private == "" ? true : fileexists(var.ssh_key_private)
    error_message = "The provided SSH key file does not exist. Please, provide a valid path."
  }
}

variable "datacenter_name" {
  type        = string
  description = "The name of the datacenter in vSphere."

  validation {
    condition     = var.deploy-vmware ? var.datacenter_name != "REPLACE ME" && var.datacenter_name != "" : true
    error_message = "Provide the correct VMware vSphere datacenter name."
  }
}

variable "folder_name" {
  type        = string
  description = "The name of the folder in vSphere."

  validation {
    condition     = var.deploy-vmware ? var.folder_name != "REPLACE ME" && var.folder_name != "" : true
    error_message = "Provide the correct VMware vSphere folder name."
  }
}

variable "search_domain" {
  type        = string
  description = "The name of network search domain."

  validation {
    condition     = var.deploy-vmware ? var.search_domain != "REPLACE ME" && var.search_domain != "" : true
    error_message = "Provide the correct VMware vSphere search domain."
  }
}

# Input resources for the cluster - Placement
variable "vsphere_cluster" {
  type        = string
  description = "The name of your vSphere cluster."

  validation {
    condition     = var.deploy-vmware ? var.vsphere_cluster != "REPLACE ME" && var.vsphere_cluster != "" : true
    error_message = "Provide the correct VMware vSphere cluster name."
  }
}

variable "datastore_name" {
  type        = string
  description = "The name of the vSphere datastore."

  validation {
    condition     = var.deploy-vmware ? var.datastore_name != "REPLACE ME" && var.datastore_name != "" : true
    error_message = "Provide the correct VMware vSphere datastore name."
  }
}

variable "network_name" {
  type        = string
  description = "The name of the vSphere network."

  validation {
    condition     = var.deploy-vmware ? var.network_name != "REPLACE ME" && var.network_name != "" : true
    error_message = "Provide the correct VMware vSphere network name."
  }
}

variable "resource_pool_name" {
  type        = string
  description = "The name of the vSphere resource pool."

  validation {
    condition     = var.deploy-vmware ? var.resource_pool_name != "REPLACE ME" && var.resource_pool_name != "" : true
    error_message = "Provide the correct VMware vSphere resource pool name."
  }
}

variable "pcg_name" {
  type        = string
  description = "The name of the PCG that will be used to deploy the cluster."

  validation {
    condition     = var.deploy-vmware ? var.pcg_name != "REPLACE ME" && var.pcg_name != "" : true
    error_message = "Provide the correct VMware vSphere PCG name."
  }
}

# Input resources for the Static IP Pool (required for static IP placement only)
variable "deploy-vmware-static" {
  type        = bool
  description = "A flag for enabling a deployment on VMware using static IP placement."
}

variable "network_gateway" {
  type        = string
  description = "The IP address of the vSphere network gateway."
}

variable "network_prefix" {
  type        = number
  description = "The prefix of your vSphere network. Valid values are network CIDR subnet masks from the range 0-32. Example: 18."
}

variable "ip_range_start" {
  type        = string
  description = "The first IP address of your PCG IP pool range."
}

variable "ip_range_end" {
  type        = string
  description = "The last IP address of your PCG IP pool range."
}

variable "nameserver_addr" {
  type        = set(string)
  description = "A comma-separated list of DNS nameserver IP addresses of your network."
}


##############################
# Hello Universe App Variables
##############################
variable "app_namespace" {
  type        = string
  description = "The namespace in which the application will be deployed."
}

variable "app_port" {
  type        = number
  description = "The cluster port number on which the service will listen for incoming traffic."
}

variable "replicas_number" {
  type        = number
  description = "The number of pods to be created."
}

variable "db_password" {
  type        = string
  description = "The base64 encoded database password to connect to the API database."
  sensitive   = true

  validation {
    condition     = var.deploy-aws || var.deploy-azure || var.deploy-gcp || var.deploy-vmware ? var.db_password != "REPLACE ME" && var.db_password != "" : true
    error_message = "Provide the correct database password."
  }
}

variable "auth_token" {
  type        = string
  description = "The base64 encoded auth token for the API connection."
  sensitive   = true

  validation {
    condition     = var.deploy-aws || var.deploy-azure || var.deploy-gcp || var.deploy-vmware ? var.auth_token != "REPLACE ME" && var.auth_token != "" : true
    error_message = "Provide the correct authentication token."
  }
}
