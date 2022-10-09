# Pack Registry
variable "registry_name" {
  default = "Public"
}

# Cluster
variable "host_cluster_uid" {}
variable "chart_name" {
  default = ""
}
variable "chart_repo" {
  default = ""
}
variable "chart_values" {
  default = ""
}
variable "chart_version" {
  default = ""
}
variable "k8s_version" {
  default = ""
}

# CI/CD
variable "docker_config" {}
variable "external_domain" {}
variable "github_access_token" {}
variable "github_org" {}
variable "github_repo" {}
variable "github_user" {}
variable "image_source" {}