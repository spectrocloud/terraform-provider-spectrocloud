# Looks up a user by email address.
data "spectrocloud_user" "example" {
  # Lookup key, optional, also Computed. The schema also exposes `id` (ConflictsWith `email`),
  # but only `email`-based lookup is actually implemented by the provider today - setting `id`
  # alone does not resolve anything; use `email` as shown here.
  email = "user@example.com"
}

# Output user details for reference
output "user_info" {
  value = {
    id    = data.spectrocloud_user.example.id
    email = data.spectrocloud_user.example.email
  }
}
