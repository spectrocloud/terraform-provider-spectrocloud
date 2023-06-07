data "spectrocloud_filter" "example" {
  name = "resourcefilter2"
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
