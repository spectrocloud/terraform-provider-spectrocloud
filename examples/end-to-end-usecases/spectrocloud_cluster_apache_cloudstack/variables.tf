variable "private_cloud_gateway_id" {
  description = "UID of the Private Cloud Gateway this cloud account connects through."
}

variable "cloudstack_api_url" {
  description = "The CloudStack management server's API endpoint."
  type        = string
}

variable "cloudstack_api_key" {
  description = "CloudStack API key."
  type        = string
  sensitive   = true
}

variable "cloudstack_secret_key" {
  description = "CloudStack secret key."
  type        = string
  sensitive   = true
}

variable "cloudstack_zone_name" {
  description = "CloudStack zone name where the cluster will be deployed."
  type        = string
}

variable "cloudstack_network_name" {
  description = "CloudStack network name within the selected zone."
  type        = string
}

variable "cloudstack_ssh_key_name" {
  description = "SSH key name (registered in CloudStack) for accessing cluster nodes."
  type        = string
}

variable "cloudstack_cp_offering" {
  description = "CloudStack compute offering (instance type/size) name for the control-plane pool."
  type        = string
}

variable "cloudstack_worker_offering" {
  description = "CloudStack compute offering (instance type/size) name for the worker pool."
  type        = string
}

variable "backup_minio_endpoint" {
  description = "The S3-compatible (Minio) endpoint URL used for backup storage, e.g. https://minio.mycompany.com."
  type        = string
}

variable "backup_minio_access_key" {
  description = "Access key for the Minio backup storage endpoint."
  type        = string
  sensitive   = true
}

variable "backup_minio_secret_key" {
  description = "Secret key for the Minio backup storage endpoint."
  type        = string
  sensitive   = true
}

variable "kubeconfig_output_dir" {
  description = "Local directory to write the cluster's kubeconfig files into."
  default     = "."
}
