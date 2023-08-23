variable "aws_access_key" {}
variable "aws_secret_key" {}

# Cluster
variable "aws_ssh_key_name" {}
variable "aws_region" {}
variable "aws_region_az" {}


# Provisioning Option B (Static)
variable "master_azs_subnets_map" {
  default = {}
  type    = map(string)
}

variable "worker_azs_subnets_map" {
  default = {}
  type    = map(string)
}