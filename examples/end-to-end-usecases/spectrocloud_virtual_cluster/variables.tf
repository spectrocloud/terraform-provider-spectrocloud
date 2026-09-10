# spectrocloud_virtual_cluster has no cloud account and provisions no infrastructure of its own -
# it's a Kubernetes-in-Kubernetes control plane (vcluster) that runs inside an already-existing
# Palette host cluster (or whichever cluster a cluster group selects), so this example has no
# cloudaccount.tf.

variable "host_cluster_uid" {
  description = "UID of the existing Palette-managed cluster this virtual cluster is created on."
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
