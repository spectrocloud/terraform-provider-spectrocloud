data "spectrocloud_private_cloud_gateway" "pcg" {
  name = var.private_cloud_gateway_name
}

# Day-2 mutability: `name` and `private_cloud_gateway_id` are ForceNew - changing either
# recreates the IP pool. `network_type`, `ip_start_range`/`ip_end_range`/`subnet_cidr`, `prefix`,
# `gateway`, `nameserver_addresses`, `nameserver_search_suffix`, and
# `restrict_to_single_cluster` all update in place.
resource "spectrocloud_privatecloudgateway_ippool" "range_example" {
  # Required, ForceNew.
  name = "ippool-range-example"
  # Required, ForceNew.
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id

  # Required. Allowed: "range" or "subnet" - selects which of the fields below is used.
  network_type = "range"
  # Required when network_type = "range". Mutually exclusive with subnet_cidr in practice.
  ip_start_range = "10.10.10.100"
  ip_end_range   = "10.10.10.200"

  # Required. Prefix length for the pool's network range, e.g. 24 for a /24.
  prefix = 24
  # Required. Network gateway IP for this pool - typically the subnet's default gateway.
  gateway = "10.10.10.1"

  # Optional.
  nameserver_addresses     = ["10.10.10.2", "10.10.10.3"]
  nameserver_search_suffix = ["example.org"]

  # Optional, default false. Recommended true for production - restricts this pool to a single
  # cluster instead of being shared across clusters.
  restrict_to_single_cluster = true
}

resource "spectrocloud_privatecloudgateway_ippool" "subnet_example" {
  name                     = "ippool-subnet-example"
  private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.pcg.id

  network_type = "subnet"
  # Required when network_type = "subnet". Mutually exclusive with ip_start_range/ip_end_range.
  subnet_cidr = "10.10.20.0/24"

  prefix  = 24
  gateway = "10.10.20.1"
}
