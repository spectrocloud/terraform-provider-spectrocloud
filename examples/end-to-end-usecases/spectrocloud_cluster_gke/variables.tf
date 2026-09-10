# --- GCP cloud account ---
variable "gcp_json_credentials" {
  description = "GCP service account credentials in JSON format (credential material)."
  sensitive   = true
}

# --- Cluster placement ---
variable "gcp_project_id" {
  description = "GCP project ID the cluster is provisioned into."
}
variable "gcp_region" {
  description = "GCP region, e.g. us-east1."
  default     = "us-east1"
}

# --- Backup storage location (GCS) ---
variable "backup_gcp_project_id" {
  description = "GCP project ID for the backup storage location (can be the same as gcp_project_id)."
}
variable "backup_gcp_json_credentials" {
  description = "GCP service account credentials in JSON format for the backup storage location."
  sensitive   = true
}

# --- Local kubeconfig export ---
variable "kubeconfig_output_dir" {
  description = "Local directory to write the cluster's kubeconfig files into."
  default     = "."
}
