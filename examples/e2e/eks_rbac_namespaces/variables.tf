variable "cloud_account_type" {}

# Option A (When Using access key and secret key)
variable "aws_access_key" {
  default = ""
}
variable "aws_secret_key" {
  default = ""
}

# Option B (When Using sts info, arn and external id)
variable "arn" {
  default = ""
}
variable "external_id" {
  default = ""
}

# Cluster
variable "aws_ssh_key_name" {
  default = ""
}
variable "aws_region" {}
variable "aws_vpc_id" {
  default = ""
}

variable "master_azs_subnets_map" {
  type = map(string)
}

variable "worker_azs_subnets_map" {
  type = map(string)
}
