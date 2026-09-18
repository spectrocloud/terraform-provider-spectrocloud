variable "oidc_client_id" {
  description = "OIDC client ID registered with the identity provider."
}

variable "oidc_client_secret" {
  description = "OIDC client secret registered with the identity provider."
  sensitive   = true
}
