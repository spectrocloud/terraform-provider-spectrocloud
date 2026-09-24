# Looks up a single Edge Native appliance (host) registered in Palette.
#
# Lookup keys (exactly one of `id`/`name` required):
#   id   - ID of the appliance in Palette.
#   name - Name of the appliance.
data "spectrocloud_appliance" "example" {
  id = "appliance-1234"
  # name = "example-appliance"
}

# Computed (read-only) outputs:
#   tags         - Tags applied to the appliance.
#   status       - One of "ready", "in-use", "unpaired".
#   health       - One of "healthy", "unhealthy".
#   architecture - One of "amd64", "arm64".
output "appliance_details" {
  value = {
    id           = data.spectrocloud_appliance.example.id
    name         = data.spectrocloud_appliance.example.name
    tags         = data.spectrocloud_appliance.example.tags
    status       = data.spectrocloud_appliance.example.status
    health       = data.spectrocloud_appliance.example.health
    architecture = data.spectrocloud_appliance.example.architecture
  }
}
