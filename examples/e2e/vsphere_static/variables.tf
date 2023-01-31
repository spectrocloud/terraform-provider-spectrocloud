#cluster variables
variable "vsphere_image_template_folder" {}
variable "cluster_ssh_public_key" {}
variable "cluster_network_search" {}

variable "vsphere_datacenter" {}
variable "vsphere_folder" {}

variable "vsphere_cluster" {}
variable "vsphere_resource_pool" {}
variable "vsphere_datastore" {}
variable "vsphere_network" {}

# common
variable "gateway_name" {
  type = string
}

variable "ippool_name" {
  type = string
}

# cloud account variables
variable "new_vmware_cloud_account_name" {}

variable "vsphere_vcenter" {
  type = string
}

variable "vsphere_username" {
  type = string
}

variable "vsphere_password" {
  type      = string
  sensitive = true
}