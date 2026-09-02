variable "maas_api_endpoint" {}
variable "maas_api_key" {}

# Cluster
variable "private_cloud_gateway_id" {}
variable "maas_resource_pool" {}
variable "maas_domain" {}
variable "maas_azs" {}

variable "cluster_ssh_public_keys" {
  description = "A list of SSH public keys to inject into MAAS nodes as authorized keys for the 'spectro' user."
  type        = list(string)
  default     = []
}
