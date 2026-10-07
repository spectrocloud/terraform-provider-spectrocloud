# spectrocloud_cluster_group has no cloud account and provisions no infrastructure of its own -
# it's a logical overlay ("hostCluster" type) over one or more already-existing Palette-managed
# clusters, so this example has no cloudaccount.tf. Two host clusters are referenced here (not
# just one) to exercise the repeatable `clusters` block.

variable "host_cluster_uid_primary" {
  description = "UID of an existing Palette-managed cluster to add to the group as a host cluster."
  type        = string
}
variable "host_cluster_host_dns_primary" {
  description = "Host DNS wildcard Palette uses to route to virtual clusters scheduled on the primary host cluster, e.g. \"*.dev.example.com\"."
  type        = string
}

variable "host_cluster_uid_secondary" {
  description = "UID of a second existing Palette-managed cluster to add to the group as a host cluster."
  type        = string
}
variable "host_cluster_host_dns_secondary" {
  description = "Host DNS wildcard for the secondary host cluster, e.g. \"*.staging.example.com\"."
  type        = string
}
