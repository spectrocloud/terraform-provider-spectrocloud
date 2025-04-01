# Data source to retrieve details of appliances based on filters
data "spectrocloud_appliances" "filtered_appliances" {
  context      = "project"        # Context can be "project" or "tenant"
  status       = "ready"         # Filter by status ready, in-use, unpaired
  health       = "healthy"        # Filter by health status
  architecture = "amd_64"         # Filter by architecture type amd64, arm64
  tags = {
    environment = "production"    # Filter by tag key-value pairs
  }
}

# Output the list of appliance IDs that match the filters
output "appliance_ids" {
  value = [for a in data.spectrocloud_appliance.filtered_appliances : a.name]
}