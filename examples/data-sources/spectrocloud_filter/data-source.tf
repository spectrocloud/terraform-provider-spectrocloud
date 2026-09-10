# Looks up a tag-based filter definition by name.
data "spectrocloud_filter" "example" {
  # Required lookup key.
  name = "example-filter"
}

# Computed. {name, annotations, labels} - use metadata[0].<field> to reach a specific field.
output "filter_metadata" {
  value = data.spectrocloud_filter.example.metadata
}

# Computed. {filter_group: [{conjunction, filters: [{key, negation, operator, values}]}]}.
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

# Computed. Typically "and" or "or", matching the values accepted by the spectrocloud_filter
# resource's own conjunction field.
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
