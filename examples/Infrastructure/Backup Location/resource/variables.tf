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

variable "aws_access_key" {
  description = "AWS access key for the S3/Minio backup storage locations using credential_type = \"secret\""
  type        = string
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret key for the S3/Minio backup storage locations using credential_type = \"secret\""
  type        = string
  sensitive   = true
}

variable "aws_sts_role_arn" {
  description = "IAM role ARN to assume for the S3 backup storage location using credential_type = \"sts\""
  type        = string
}

variable "aws_external_id" {
  description = "External ID for cross-account STS role assumption on the S3 backup storage location"
  type        = string
}

variable "gcp_json_credentials" {
  description = "GCP service account credentials, in JSON format, for the GCP backup storage location"
  type        = string
  sensitive   = true
}

variable "azure_client_secret" {
  description = "Azure client secret for the Azure backup storage location"
  type        = string
  sensitive   = true
}
