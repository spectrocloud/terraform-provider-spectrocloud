variable "control_plane_edge_host_uid" {
  description = "UID of the physical/virtual edge appliance to pair as the control-plane node. Obtained by pairing the appliance in Palette (Clusters > Edge Hosts) with the pairing_key shown there."
  type        = string
}

variable "worker_edge_host_uid" {
  description = "UID of the physical/virtual edge appliance to pair as the worker node."
  type        = string
}

variable "cluster_ssh_public_keys" {
  description = "SSH public keys injected into the cluster nodes as authorized keys."
  type        = list(string)
  default     = []
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
