# Retrieve details of a specific application profile
data "spectrocloud_application_profile" "example_profile" {
  name = "my-app-profile"  # Specify the name of the application profile
}

# Output the retrieved application profile details
output "application_profile_version" {
  value = data.spectrocloud_application_profile.example_profile.version
}