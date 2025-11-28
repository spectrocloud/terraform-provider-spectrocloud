variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  type        = string
  sensitive   = true
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

# Cluster Configuration
variable "cluster_name" {
  description = "Name of the Apache CloudStack cluster"
  type        = string
}

variable "cluster_cloud_account_name" {
  description = "Name of the Apache CloudStack cloud account"
  type        = string
}

variable "cluster_cluster_profile_name" {
  description = "Name of the cluster profile to use"
  type        = string
}

# CloudStack Configuration
variable "cloudstack_zone_name" {
  description = "CloudStack zone name where the cluster will be deployed"
  type        = string
}

variable "cloudstack_network_name" {
  description = "CloudStack network name for the cluster"
  type        = string
}

variable "cloudstack_compute_offering" {
  description = "CloudStack compute offering for control plane nodes"
  type        = string
}

variable "cloudstack_compute_offering_worker" {
  description = "CloudStack compute offering for worker nodes"
  type        = string
}

variable "ssh_key_name" {
  description = "SSH key name for cluster node access (optional)"
  type        = string
  default     = ""
}

# Optional: Static IP Pool
variable "static_ip_pool_id" {
  description = "Static IP pool ID for cluster nodes (optional)"
  type        = string
  default     = ""
}

# Optional: Backup Storage Location
variable "backup_storage_location_name" {
  description = "Name of the backup storage location (optional)"
  type        = string
  default     = ""
}

