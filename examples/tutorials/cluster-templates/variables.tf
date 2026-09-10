#########
# Palette
#########

variable "palette-project" {
  description = "Palette project name"
  type        = string

  validation {
    condition     = var.palette-project != ""
    error_message = "Provide the correct Palette project."
  }
}

#####################
# Hello Universe
#####################

variable "app_port" {
  type        = number
  description = "The cluster port number on which the service will listen for incoming traffic."
}

variable "create_new_profile_version" {
  type        = bool
  description = "Create cluster profile version 1.1.0 with Kubecost added. Set to true after the initial clusters are deployed."
  default     = false
}

variable "update_template_profile_version" {
  type        = bool
  description = "Update the cluster template to reference cluster profile version 1.1.0. Requires create_new_profile_version to be true and already applied."
  default     = false
}

variable "upgrade_now_timestamp" {
  type        = string
  description = "RFC3339 timestamp that triggers an immediate upgrade on all template clusters when changed. Set a new timestamp to re-trigger the upgrade. Leave empty to skip."
  default     = ""
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

variable "aws-region" {
  type        = string
  description = "AWS region."
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

#########
# Azure
#########

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
    instance_type       = "Standard_D4s_v5"
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
    instance_type       = "Standard_D4s_v5"
    disk_size_gb        = "60"
    azs                 = ["1"]
    is_system_node_pool = false
  }
  description = "Azure worker nodes configuration."
}
