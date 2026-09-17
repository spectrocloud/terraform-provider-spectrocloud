# Looks up the set of Edge Native appliances (hosts) matching the given filters. Every
# attribute below except `ids` is a filter you set; `ids` is the sole computed output - the
# result does not include the appliances' other properties (name, tags, etc.), only their IDs.
#
# Filters (all optional):
#   context      - default "project". Allowed: "project", "tenant".
#   status       - Allowed: "ready", "in-use", "unpaired". Omit to match any status.
#   health       - Allowed: "healthy", "unhealthy". Omit to match any health state.
#   architecture - Allowed: "amd64", "arm64". Omit to match any architecture.
#   tags         - Matches appliances carrying all of these tag key/value pairs.
data "spectrocloud_appliances" "filtered_appliances" {
  context      = "project"
  status       = "ready"
  health       = "healthy"
  architecture = "amd64"
  tags = {
    environment = "production"
  }
}

# Computed. IDs of every appliance matching the filters above.
output "appliance_ids" {
  value = data.spectrocloud_appliances.filtered_appliances.ids
}
