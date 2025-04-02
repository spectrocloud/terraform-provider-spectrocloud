data "provider_appliance" "example" {
  # You can specify either `id` or `name`, but not both.
  id = "appliance-1234"
  # name = "example-appliance"
}

output "appliance_details" {
  value = {
    id           = data.provider_appliance.example.id
    name         = data.provider_appliance.example.name
    tags         = data.provider_appliance.example.tags
    status       = data.provider_appliance.example.status
    health       = data.provider_appliance.example.health
    architecture = data.provider_appliance.example.architecture
  }
}
