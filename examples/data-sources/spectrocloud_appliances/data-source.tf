# Looks up the set of Edge Native appliances (hosts) matching the given filters. Every
# attribute below except `ids` is a filter you set; `ids` is the sole computed output - the
# result does not include the appliances' other properties (name, tags, etc.), only their IDs.
data "spectrocloud_appliances" "filtered_appliances" {
  # Lookup filter, optional, default "project". Allowed: "project", "tenant".
  context = "project"
  # Lookup filter, optional. Allowed: "ready", "in-use", "unpaired". Omit to match any status.
  status = "ready"
  # Lookup filter, optional. Allowed: "healthy", "unhealthy". Omit to match any health state.
  health = "healthy"
  # Lookup filter, optional. Allowed: "amd64", "arm64". Omit to match any architecture.
  architecture = "amd64"
  # Lookup filter, optional. Matches appliances carrying all of these tag key/value pairs.
  tags = {
    environment = "production"
  }
}

# Computed. IDs of every appliance matching the filters above.
output "appliance_ids" {
  value = data.spectrocloud_appliances.filtered_appliances.ids
}
