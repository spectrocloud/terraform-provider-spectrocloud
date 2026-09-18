variable "sc_host" {
  description = "Spectro Cloud endpoint"
  default     = "api.spectrocloud.com"
}

variable "sc_api_key" {
  description = "Spectro Cloud API key"
  type        = string
  sensitive   = true
}

variable "aws_access_key" {
  description = "AWS access key for CloudWatch audit trail secret credentials"
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret key for CloudWatch audit trail secret credentials"
  type        = string
  sensitive   = true
}

variable "aws_sts_role_arn" {
  description = "AWS IAM role ARN for CloudWatch audit trail STS credentials"
  type        = string
  default     = ""
}

variable "aws_external_id" {
  description = "External ID for STS role assumption"
  type        = string
  sensitive   = true
  default     = ""
}

variable "splunk_hec_token" {
  description = "Splunk HEC token for Splunk audit trail"
  type        = string
  sensitive   = true
  default     = ""
}
