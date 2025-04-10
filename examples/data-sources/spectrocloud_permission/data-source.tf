# Fetches details of a specific permission in SpectroCloud
data "spectrocloud_permission" "example" {
  # The name of the permission (Required)
  # Example: "App Deployment", "Cluster Management", "User Access"
  name = "App Deployment"

  # Scope of the permission (Optional, Defaults to "project")
  # Allowed values: "project", "tenant", "resource"
  scope = "project"
}

# Output the retrieved permission details
output "permission_details" {
  value = data.spectrocloud_permission.example
}

# Individual outputs for better clarity (optional)
output "permission_id" {
  value = data.spectrocloud_permission.example.id
}

output "permission_list" {
  value = data.spectrocloud_permission.example.permissions
}