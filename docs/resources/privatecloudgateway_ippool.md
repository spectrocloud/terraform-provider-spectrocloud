---
page_title: "spectrocloud_privatecloudgateway_ippool Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# Resource `spectrocloud_privatecloudgateway_ippool`



## Example Usage

```terraform
resource "spectrocloud_privatecloudgateway_ippool" "ippool" {
  name                 = "ippool-dev"
  private_cloud_gateway_id = "5fbb9195f4a3d7b95f888211"
  network_type = "range"
  ip_start_range = "10.10.10.100"
  ip_end_range = "10.10.10.199"
  prefix = 24
  gateway = "10.10.10.1"
  nameserver_addresses = "8.8.8.8"
  nameserver_search_suffix = "test.com"
  restrict_to_single_cluster = false
}

resource "spectrocloud_privatecloudgateway_ippool" "ippoolprod" {
  name                 = "ippool-prod"
  private_cloud_gateway_id = "5fbb9195f4a3d7b95f888211"
  network_type = "subnet"
  subnet_cidr = "10.10.10.100/16"
  prefix = 30
  gateway = "10.10.10.1"
  nameserver_addresses = "8.8.8.8"
  nameserver_search_suffix = "test.com"
  restrict_to_single_cluster = true
}
```

## Schema

### Required

- **name** (String)
- **private_cloud_gateway_id** (String)
- **network_type** (String) one of [`range`, `subnet`]
- **prefix** (Int)
- **gateway** (String)
- **ip_start_range** (String) if `network_type` is `range`
- **ip_end_range** (String) if `network_type` is `range`
- **subnet_cidr** (String) if `network_type` is `subnet`

### Optional

- **nameserver_addresses** (String) Comma seperated value
- **nameserver_search_suffix** (String) Comma seperated value
- **restrict_to_single_cluster** (Boolean)
