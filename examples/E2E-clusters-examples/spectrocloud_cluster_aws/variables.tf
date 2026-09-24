# --- AWS cloud account ---
variable "aws_access_key" {
  description = "AWS access key for the cloud account (credential material)."
  sensitive   = true
}
variable "aws_secret_key" {
  description = "AWS secret key for the cloud account (credential material)."
  sensitive   = true
}

# --- Cluster placement ---
variable "cluster_ssh_key_name" {
  description = "Name of an existing EC2 key pair to use for cluster nodes."
}
variable "aws_region" {
  description = "AWS region to deploy the cluster in."
  default     = "us-east-1"
}

# --- Backup storage location (S3) ---
variable "backup_s3_bucket_name" {
  description = "Name of the S3 bucket backups are written to."
}
variable "backup_s3_access_key" {
  description = "AWS access key for the backup storage location (credential material)."
  sensitive   = true
}
variable "backup_s3_secret_key" {
  description = "AWS secret key for the backup storage location (credential material)."
  sensitive   = true
}

# --- Local kubeconfig export ---
variable "kubeconfig_output_dir" {
  description = "Local directory to write the cluster's kubeconfig files into."
  default     = "."
}
