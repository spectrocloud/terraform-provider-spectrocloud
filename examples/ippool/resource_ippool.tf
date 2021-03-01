resource "spectrocloud_privatecloudgateway_ippool" "ippool1" {
  name                 = "ippool-1"
  private_cloud_gateway_id = "5fbb9195f4a3d7b95f888211"
  network_type = "range"
  ip_start_range = "10.12.10.120"
  ip_end_range = "10.12.10.125"
  prefix = 24
  gateway = "10.12.10.1"
  nameserver_addresses = ["8.8.8.8"]
  nameserver_search_suffix = ["test.com"]
  restrict_to_single_cluster = true
}

resource "spectrocloud_privatecloudgateway_ippool" "ippool2" {
  name                 = "ippool-2"
  private_cloud_gateway_id = "5fbb9195f4a3d7b95f888211"
  network_type = "subnet"
  subnet_cidr = "100.12.10.120/16"
  prefix = 30
  gateway = "100.12.10.1"
  nameserver_addresses = ["8.8.8.8"]
  nameserver_search_suffix = ["test.com"]
  restrict_to_single_cluster = true
}