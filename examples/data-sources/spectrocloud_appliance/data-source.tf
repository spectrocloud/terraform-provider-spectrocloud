# Looks up a single Edge Native appliance (host) registered in Palette.
data "spectrocloud_appliance" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required). ID of the appliance in Palette.
  id = "appliance-1234"
  # Lookup key, optional (exactly one of `id`/`name` required). Name of the appliance.
  # name = "example-appliance"
}

output "appliance_details" {
  value = {
    id   = data.spectrocloud_appliance.example.id
    name = data.spectrocloud_appliance.example.name
    # Computed. Tags applied to the appliance.
    tags = data.spectrocloud_appliance.example.tags
    # Computed. One of "ready", "in-use", "unpaired".
    status = data.spectrocloud_appliance.example.status
    # Computed. One of "healthy", "unhealthy".
    health = data.spectrocloud_appliance.example.health
    # Computed. One of "amd64", "arm64".
    architecture = data.spectrocloud_appliance.example.architecture
  }
}
