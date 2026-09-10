# Looks up an existing application profile by name (and, optionally, version).
data "spectrocloud_application_profile" "example_profile" {
  # Required lookup key. Name of the application profile.
  name = "my-app-profile"
  # Optional lookup key, also Computed. Set to look up a specific version; if omitted, looks up
  # "1.0.0" and returns whichever version was actually matched.
  # version = "1.0.0"
}

output "application_profile_version" {
  value = data.spectrocloud_application_profile.example_profile.version
}
