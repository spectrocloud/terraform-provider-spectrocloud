# Cluster
variable "cluster_ssh_key_name" {
  default = "spectro2022"
}

variable "cluster_network_search" {}

variable "vsphere_datacenter" {}
variable "vsphere_folder" {}

variable "vsphere_cluster" {}
variable "vsphere_resource_pool" {}
variable "vsphere_datastore" {}
variable "vsphere_network" {}
