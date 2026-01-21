variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  type        = string
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  type        = string
  sensitive   = true
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  type        = string
  default     = "Default"
}



variable "kubeconfig_path" {
  description = "Path to kubeconfig file (optional, defaults to ~/.kube/config or KUBECONFIG env var)"
  type        = string
  default     = ""
}

variable "wait_timeout_seconds" {
  description = "Maximum time to wait for cluster to reach Running-Healthy state (in seconds)"
  type        = number
  default     = 300  # 5 minutes
}

variable "poll_interval_seconds" {
  description = "Interval between status checks (in seconds)"
  type        = number
  default     = 30
}
