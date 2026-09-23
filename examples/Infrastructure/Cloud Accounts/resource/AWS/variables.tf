variable "sc_host" {
  description = "Spectro Cloud Endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
}

variable "sc_project_name" {
  description = "Spectro Cloud Project (e.g: Default)"
  default     = "Default"
}

# Variables for Secret Credentials example
variable "aws_secured_access_key" {
  description = "AWS Access Key for secret credentials"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS Secret Key for secret credentials"
  type        = string
  sensitive   = true
}

# Variables for STS example
variable "aws_sts_role_arn" {
  description = "AWS IAM Role ARN for STS authentication"
  type        = string
  default     = ""
}

variable "aws_external_id" {
  description = "External ID for STS role assumption"
  type        = string
  sensitive   = true
  default     = ""
}

# Variables for Pod Identity example
variable "aws_pod_identity_role_arn" {
  description = "AWS IAM Role ARN for EKS Pod Identity"
  type        = string
  default     = ""
}

variable "aws_permission_boundary_arn" {
  description = "Permission Boundary ARN for EKS Pod Identity (optional)"
  type        = string
  default     = ""
}
