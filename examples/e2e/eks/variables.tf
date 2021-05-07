variable "aws_access_key" {}
variable "aws_secret_key" {}
variable "arn" {}
variable "external_id" {}

# Cluster
variable "aws_ssh_key_name" {}
variable "aws_region" {}

variable "cloud_account_type" {}

variable "aws_region_az" {
  type    = list(string)
}

variable "master_azs_subnets_map" {
  type = "map"
}

variable "worker_azs_subnets_map" {
  type = "map"
}