variable "edge_host_uid" {
  description = "UID of the physical/virtual edge appliance (running the vSphere-compatible virtualization layer) this cluster is deployed on. Obtained by pairing the appliance in Palette (Clusters > Edge Hosts) with the pairing_key shown there."
  type        = string
}

variable "cluster_ssh_public_key" {
  description = "SSH public key injected into the cluster nodes as an authorized key."
  type        = string
}

variable "vsphere_datacenter" {
  description = "vSphere datacenter used by this cluster."
  type        = string
}

variable "vsphere_folder" {
  description = "vSphere inventory folder where cluster virtual machines are created."
  type        = string
}

variable "vsphere_cluster" {
  description = "vSphere compute cluster used for node placement."
  type        = string
}

variable "vsphere_resource_pool" {
  description = "vSphere resource pool used for node placement."
  type        = string
}

variable "vsphere_datastore" {
  description = "vSphere datastore used for node disks."
  type        = string
}

variable "vsphere_network" {
  description = "vSphere network attached to machine pool nodes."
  type        = string
}

variable "cluster_vip" {
  description = "Virtual IP address for the Kubernetes control plane endpoint."
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
