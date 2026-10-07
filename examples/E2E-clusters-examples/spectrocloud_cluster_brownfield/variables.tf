# spectrocloud_cluster_brownfield registers an already-running Kubernetes cluster with Palette -
# there is no cloud account and no infrastructure provisioning involved, so this example has no
# cloudaccount.tf. Applying this configuration only registers the cluster's *intent* with
# Palette; the import itself completes only once the generated kubectl_command (see outputs.tf
# and import_command.tf) is run against the real, already-existing cluster.

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

variable "import_command_output_dir" {
  description = "Local directory to write the generated kubectl import command into."
  default     = "."
}
