variable "cloud_account_type" {}

# Option A (When Using access key and secret key)
variable "aws_access_key" {}
variable "aws_secret_key" {}

# Option B (When Using sts info, arn and external id)
variable "arn" {}
variable "external_id" {}

# Cluster
variable "aws_ssh_key_name" {}
variable "aws_region" {}

variable "aws_region_az" {
  type    = list(string)
}

variable "master_azs_subnets_map" {
  type = "map"
}

variable "worker_azs_subnets_map" {
  type = "map"
}