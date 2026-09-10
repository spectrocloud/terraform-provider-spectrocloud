variable "edge_host_uid" {
  description = "UID of the registered Edge host where this cluster is deployed."
}

variable "cluster_ssh_public_key" {
  description = "Public SSH key used to access the cluster nodes."
}

variable "vsphere_datacenter" {}
variable "vsphere_folder" {}
variable "vsphere_cluster" {}
variable "vsphere_resource_pool" {}
variable "vsphere_datastore" {}
variable "vsphere_network" {}

variable "cluster_vip" {
  description = "Virtual IP address for the Kubernetes control plane endpoint."
}

variable "backup_storage_location_name" {}
