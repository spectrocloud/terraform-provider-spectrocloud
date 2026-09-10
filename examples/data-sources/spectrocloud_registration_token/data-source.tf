# Looks up an Edge registration token by name or by UID.
data "spectrocloud_registration_token" "tf" {
  # Lookup key, optional (at least one of `name`/`id` required).
  name = "ran-dev-test"
  # Lookup key, optional (conflicts with `name`), also Computed.
  # id = "657ec9a27afca71b0dc98027"
}

# Computed, sensitive. Treat as a credential.
output "token" {
  value     = data.spectrocloud_registration_token.tf.token
  sensitive = true
}

# Computed.
output "project_uid" {
  value = data.spectrocloud_registration_token.tf.project_uid
}

output "expiry_date" {
  value = data.spectrocloud_registration_token.tf.expiry_date
}

# Computed. "active" or "inactive".
output "status" {
  value = data.spectrocloud_registration_token.tf.status
}
