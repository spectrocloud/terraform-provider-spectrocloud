####################################
# Input resources for the profile
####################################

variable "cluster_profile_name" {
  type        = string
  description = "The name of the cluster profile."
  default     = "pcg-tutorial-profile"
}

variable "cluster_profile_description" {
  type        = string
  description = "Provide a description of the cluster profile."
  default     = "My cluster profile as part of the PCG tutorial."
}

variable "metallb_ip" {
  type        = string
  description = "The IP address range for your MetalLB Load Balancer. This range must be included in the PCG's static IP pool range if using static IP placement."
}

####################################
# Input resources for the cluster
####################################

variable "cluster_name" {
  type        = string
  description = "The name of the cluster."
  default     = "pcg-tutorial-cluster"
}

variable "tags" {
  type        = list(string)
  description = "The default tags to apply to Palette resources. Each tag must be 63 characters or fewer, start and end with an alphanumeric character, and contain only alphanumeric characters, dots, dashes, or underscores (no slashes)."
  default     = ["spectro-cloud-education", "app:hello-universe", "terraform_managed:true", "repository:spectrocloud:tutorials", "tutorial:DEPLOY_APP_WORKLOADS_WITH_A_PCG"]
}

#################################################
# Input resources for the cluster - Cloud config
#################################################

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
}

variable "folder_name" {
  type        = string
  description = "The name of the folder in vSphere."
}

variable "search_domain" {
  type        = string
  description = "The name of the network search domain."
}

#################################################
# Input resources for the cluster - Placement
#################################################

variable "vsphere_cluster" {
  type        = string
  description = "The name of your vSphere cluster."
}

variable "datastore_name" {
  type        = string
  description = "The name of the vSphere datastore."
}

variable "network_name" {
  type        = string
  description = "The name of the vSphere network."
}

variable "resource_pool_name" {
  type        = string
  description = "The name of the vSphere resource pool."
}

variable "pcg_name" {
  type        = string
  description = "The name of the Private Cloud Gateway (PCG) that will be used to deploy the cluster. The PCG itself must already be installed — see the README prerequisites."
}
