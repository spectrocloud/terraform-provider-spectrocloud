
# SpectroCloud authentication variables
variable "sc_host" {}
variable "sc_api_key" {}
variable "sc_project_name" {}

# Cluster Config related
variable "aws_ssh_key_name" {
  default     = "spectro2020"
  description = "The SSH key to use for cluster provisioning"
}
variable "control_plane_lb" {
  default     = ""  # [ Use `internal` for private API server]
  description = "The ControlPlane API Server LoadBalancer type to use for cluster provisioning"
}

# Addon pack variables
variable "argocd_name" {
  default     = "argo-cd"
  description = "ArgoCD pack name to use for the profile & cluster"
}
variable "argocd_version" {
  default     = "3.26.7"
  description = "ArgoCD pack version to use for the profile & cluster"
}
