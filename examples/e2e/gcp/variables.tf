variable "region" {}

variable "gcp-cloud-account-name" {
  type        = string
  description = "The name of your GCP account as assigned in Palette"
}

variable "cp_nodes" {
  type = object({
    count              = string
    instance_type      = string
    disk_size_gb       = string
    availability_zones = list(string)
  })
  description = "Control Plane nodes configuration."
}

variable "worker_nodes" {
  type = object({
    count              = string
    instance_type      = string
    disk_size_gb       = string
    availability_zones = list(string)
  })
  description = "Worker nodes configuration."
}

