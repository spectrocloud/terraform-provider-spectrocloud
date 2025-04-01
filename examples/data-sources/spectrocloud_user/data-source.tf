# Fetch details of a specific user in SpectroCloud
data "spectrocloud_user" "example" {
  # Provide either `id` or `email`, but not both.
  id = "user-12345"
  # email = "user@example.com"  # Alternative way to reference a user by email
}

# Output user details for reference
output "user_info" {
  value = {
    id    = data.spectrocloud_user.example.id
    email = data.spectrocloud_user.example.email
  }
}