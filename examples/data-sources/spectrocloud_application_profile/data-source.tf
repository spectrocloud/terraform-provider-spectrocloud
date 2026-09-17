# Looks up an existing application profile by name (and, optionally, version).
#
# Lookup keys:
#   name    - Required. Name of the application profile.
#   version - Optional, also Computed. Set to look up a specific version; if omitted, looks up
#             "1.0.0" and returns whichever version was actually matched.
data "spectrocloud_application_profile" "example_profile" {
  name = "my-app-profile"
  # version = "1.0.0"
}

output "application_profile_version" {
  value = data.spectrocloud_application_profile.example_profile.version
}
