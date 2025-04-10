# Retrieve details of a specific filter by name
data "spectrocloud_filter" "example" {
  name = "example-filter"
}

# Output filter metadata for reference
output "filter_metadata" {
  value = data.spectrocloud_filter.example.metadata
}

# Output filter spec details
output "filter_spec" {
  value = data.spectrocloud_filter.example.spec
}

output "filter_name" {
  value = data.spectrocloud_filter.example.metadata[0].name
}

output "filter_annotations" {
  value = data.spectrocloud_filter.example.metadata[0].annotations
}

output "filter_labels" {
  value = data.spectrocloud_filter.example.metadata[0].labels
}

output "filter_group_conjunction" {
  value = data.spectrocloud_filter.example.spec[0].filter_group[0].conjunction
}

output "first_filter_in_group_key" {
  value = data.spectrocloud_filter.example.spec[0].filter_group[0].filters[0].key
}

output "first_filter_in_group_operator" {
  value = data.spectrocloud_filter.example.spec[0].filter_group[0].filters[0].operator
}

output "first_filter_in_group_values" {
  value = data.spectrocloud_filter.example.spec[0].filter_group[0].filters[0].values
}
