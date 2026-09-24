data "spectrocloud_private_cloud_gateway" "pcg" {
  name = var.private_cloud_gateway_name
}

# Day-2 mutability: `name` and `private_cloud_gateway_id` are ForceNew - changing either
# recreates the IP pool. `network_type`, `ip_start_range`/`ip_end_range`/`subnet_cidr`, `prefix`,
# `gateway`, `nameserver_addresses`, `nameserver_search_suffix`, and
# `restrict_to_single_cluster` all update in place.
#
# Attributes:
#   name                        - Required, ForceNew.
#   private_cloud_gateway_id    - Required, ForceNew.
#   network_type                - Required. Allowed: "range" or "subnet" - selects which of the
#                                  fields below is used.
#   ip_start_range/ip_end_range - Required when network_type = "range". Mutually exclusive with
#                                  subnet_cidr in practice.
#   subnet_cidr                 - Required when network_type = "subnet". Mutually exclusive with
#                                  ip_start_range/ip_end_range.
#   prefix                      - Required. Prefix length for the pool's network range, e.g. 24
#                                  for a /24.
#   gateway                     - Required. Network gateway IP for this pool - typically the
#                                  subnet's default gateway.
#   nameserver_addresses        - Optional.
#   nameserver_search_suffix    - Optional.
#   restrict_to_single_cluster  - Optional, default false. Recommended true for production -
#                                  restricts this pool to a single cluster instead of being shared
#                                  across clusters.
resource "spectrocloud_privatecloudgateway_ippool" "range_example" {
  name                     = "ippool-range-example"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id

  network_type   = "range"
  ip_start_range = "10.10.10.100"
  ip_end_range   = "10.10.10.200"

  prefix  = 24
  gateway = "10.10.10.1"

  nameserver_addresses     = ["10.10.10.2", "10.10.10.3"]
  nameserver_search_suffix = ["example.org"]

  restrict_to_single_cluster = true
}

resource "spectrocloud_privatecloudgateway_ippool" "subnet_example" {
  name                     = "ippool-subnet-example"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id

  network_type = "subnet"
  subnet_cidr  = "10.10.20.0/24"

  prefix  = 24
  gateway = "10.10.20.1"
}
