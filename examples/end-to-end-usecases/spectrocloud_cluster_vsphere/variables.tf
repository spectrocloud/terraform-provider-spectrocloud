# --- vSphere cloud account ---
variable "private_cloud_gateway_id" {
  description = "UID of the Private Cloud Gateway this cloud account connects through."
}
variable "vsphere_vcenter" {
  description = "vCenter server address, e.g. vcenter.corp.example.com."
}
variable "vsphere_username" {
  description = "vCenter username used to create/manage clusters."
}
variable "vsphere_password" {
  description = "vCenter password (credential material)."
  sensitive   = true
}
variable "vsphere_ignore_insecure_error" {
  description = "Accept vCenter's TLS certificate even if validation fails (self-signed certs)."
  type        = bool
  default     = false
}

# --- Cluster/machine pool placement ---
variable "cluster_ssh_public_key" {
  description = "Public SSH key injected into all cluster nodes."
}

variable "vsphere_datacenter" {}
variable "vsphere_folder" {}
variable "vsphere_cluster" {}
variable "vsphere_resource_pool" {}
variable "vsphere_datastore" {}
variable "vsphere_network" {}

# --- Backup storage location (S3) ---
variable "backup_s3_region" {
  description = "AWS region the backup S3 bucket lives in."
}
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
