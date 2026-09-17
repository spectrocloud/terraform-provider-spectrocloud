# Looks up a user by email address.
data "spectrocloud_user" "example" {
  # Lookup key, optional, also Computed, ConflictsWith `id`. Set either `email` (shown here) or
  # `id` - both are independently implemented lookups (dataSourceUserRead tries `email` first,
  # then falls back to `id`), so `id` alone works too if you already know the user's UID.
  email = "user@example.com"
}

# Equivalent lookup by ID instead of email:
# data "spectrocloud_user" "by_id" {
#   id = "64f1a2b3c4d5e6f7a8b9c0d1"
# }

# Output user details for reference
output "user_info" {
  value = {
    id    = data.spectrocloud_user.example.id
    email = data.spectrocloud_user.example.email
  }
}
