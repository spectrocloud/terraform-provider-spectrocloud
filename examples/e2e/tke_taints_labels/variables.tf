variable "tencent_secret_id" {
  default = ""
}
variable "tencent_secret_key" {
  default = ""
}

variable "tke_ssh_key_name" {
  default = ""
}
variable "tke_region" {}
variable "tke_vpc_id" {
  default = ""
}

variable "cp_tke_subnets_map" {
  type = map(string)
}

variable "worker_tke_subnets_map" {
  type = map(string)
}
