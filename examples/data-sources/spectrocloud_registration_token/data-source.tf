# Looks up an Edge registration token by name or by UID.
#
# Lookup keys (at least one of `name`/`id` required):
#   name - Optional.
#   id   - Optional (conflicts with `name`), also Computed.
data "spectrocloud_registration_token" "tf" {
  name = "ran-dev-test"
  # id = "657ec9a27afca71b0dc98027"
}

# Computed outputs:
#   token       - Sensitive. Treat as a credential.
#   project_uid
#   expiry_date
#   status      - "active" or "inactive".
output "token" {
  value     = data.spectrocloud_registration_token.tf.token
  sensitive = true
}

output "project_uid" {
  value = data.spectrocloud_registration_token.tf.project_uid
}

output "expiry_date" {
  value = data.spectrocloud_registration_token.tf.expiry_date
}

output "status" {
  value = data.spectrocloud_registration_token.tf.status
}
