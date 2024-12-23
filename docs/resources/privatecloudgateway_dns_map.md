---
page_title: "spectrocloud_privatecloudgateway_dns_map Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  This resource allows for the management of DNS mappings for private cloud gateways. This helps ensure proper DNS resolution for resources within the private cloud environment.
---

# spectrocloud_privatecloudgateway_dns_map (Resource)

  This resource allows for the management of DNS mappings for private cloud gateways. This helps ensure proper DNS resolution for resources within the private cloud environment.

You can learn more about Private Cloud Gateways DNS Mapping by reviewing the [Create and Manage DNS Mappings](https://docs.spectrocloud.com/clusters/pcg/manage-pcg/add-dns-mapping/) guide.

## Example Usage

An example of creating an DNS Map for a Private Cloud Gateway using a search domain, datacenter and network.

```hcl
 data "spectrocloud_private_cloud_gateway" "gateway" {
   name = "test-vm-pcg"
 }
 resource "spectrocloud_privatecloudgateway_dns_map" "dns_map_test" {
   private_cloud_gateway_id = data.spectrocloud_private_cloud_gateway.gateway.id
   search_domain_name = "test1.spectro.com"
   data_center = "DataCenterTest"
   network = "TEST-VM-NETWORK"
 }
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data_center` (String) The data center in which the private cloud resides.
- `network` (String) The network to which the private cloud gateway is mapped.
- `private_cloud_gateway_id` (String) The ID of the Private Cloud Gateway.
- `search_domain_name` (String) The domain name used for DNS search queries within the private cloud.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)