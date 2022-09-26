# Cluster
variable "resource_pool" {}
variable "control_plane_endpoint_url" {
  default = "test"
}
variable "control_plane_endpoint_port" {
  default = 443
}
variable "helm_release" {
  default = "test"
}
variable "k8s_version" {
  default = "1.23.0"
}