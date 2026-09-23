# spectrocloud_cluster_custom_cloud is a generic escape hatch for cloud providers that don't
# have a first-class spectrocloud_cluster_<cloud> resource - the cluster is described almost
# entirely as raw Cluster API YAML instead of structured attributes. This example uses Nutanix
# (the reference custom cloud provider shown throughout this provider's own test suite), but the
# same pattern applies to any custom cloud provider registered in Palette.

variable "private_cloud_gateway_id" {
  description = "UID of the Private Cloud Gateway (PCG) this cloud account connects through."
}

variable "nutanix_user" {
  description = "Nutanix Prism Central username."
  type        = string
}

variable "nutanix_password" {
  description = "Nutanix Prism Central password."
  type        = string
  sensitive   = true
}

variable "nutanix_endpoint" {
  description = "Nutanix Prism Central endpoint (host or IP)."
  type        = string
}

variable "nutanix_port" {
  description = "Nutanix Prism Central API port."
  default     = "9440"
}

variable "control_plane_endpoint_ip" {
  description = "Virtual IP for the Kubernetes control plane endpoint."
  type        = string
}

variable "nutanix_ssh_authorized_key" {
  description = "SSH public key injected into cluster nodes as an authorized key."
  type        = string
}

variable "nutanix_prism_element_cluster_name" {
  description = "Name of the target Nutanix Prism Element cluster."
  type        = string
}

variable "nutanix_machine_template_image_name" {
  description = "Name of the VM image template used for node provisioning."
  type        = string
}

variable "nutanix_subnet_name" {
  description = "Name of the Nutanix subnet attached to cluster nodes."
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
