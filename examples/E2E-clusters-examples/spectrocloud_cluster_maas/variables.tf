variable "private_cloud_gateway_id" {
  description = "UID of the Private Cloud Gateway (PCG) this cloud account connects through."
}

variable "maas_api_endpoint" {
  description = "The MAAS API endpoint, e.g. http://maas:5240/MAAS."
  type        = string
}

variable "maas_api_key" {
  description = "The MAAS API key."
  type        = string
  sensitive   = true
}

variable "cluster_ssh_public_keys" {
  description = "SSH public keys injected into MAAS nodes as authorized keys for the 'spectro' user."
  type        = list(string)
  default     = []
}

variable "maas_domain" {
  description = "The MAAS domain in which the cluster is provisioned."
  default     = "maas.mycompany.com"
}

variable "maas_resource_pool" {
  description = "The name of the MAAS resource pool used to place cluster nodes."
  default     = "Medium-Generic"
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
