variable "region" {}

variable "gcp-cloud-account-name" {
  type = string
  description = "The name of your GCP account as assigned in Palette"
}

variable "master_nodes" {
  type = object({
    count           = string
    instance_type   = string
    disk_size_gb    = string
    availability_zones = list(string)
  })
  description = "Master nodes configuration."
}

variable "worker_nodes" {
  type = object({
    count           = string
    instance_type   = string
    disk_size_gb    = string
    availability_zones = list(string)
  })
  description = "Worker nodes configuration."
}

