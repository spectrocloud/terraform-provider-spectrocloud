---
page_title: "spectrocloud_registry_oci Resource - terraform-provider-spectrocloud"
subcategory: ""
description: |-
  
---

# spectrocloud_registry_oci (Resource)

  

## Example Usage

```terraform
resource "spectrocloud_registry_oci" "r1" {
  name       = "test-nik2"
  type       = "ecr" # basic
  endpoint   = "123456.dkr.ecr.us-west-1.amazonaws.com"
  is_private = true
  credentials {
    credential_type = "sts"
    arn             = "arn:aws:iam::123456:role/stage-demo-ecr"
    external_id     = "sofiwhgowbrgiornM="
  }
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `credentials` (Block List, Min: 1, Max: 1) Authentication credentials to access the private OCI registry. Required if `is_private` is set to `true` (see [below for nested schema](#nestedblock--credentials))
- `endpoint` (String) The URL endpoint of the OCI registry. This is where the container images are hosted and accessed.
- `is_private` (Boolean) Specifies whether the registry is private or public. Private registries require authentication to access.
- `name` (String) The name of the OCI registry.
- `type` (String) The type of the registry. Possible values are 'ecr' (Amazon Elastic Container Registry) or 'basic' (for other types of OCI registries).

### Optional

- `base_content_path` (String) The relative path to the endpoint specified.
- `endpoint_suffix` (String) Specifies a suffix to append to the endpoint. This field is optional, but some registries (e.g., JFrog) may require it. The final registry URL is constructed by appending this suffix to the endpoint.
- `is_synchronization` (Boolean) Specifies whether the registry is synchronized.
- `provider_type` (String) The type of provider used for interacting with the registry. Supported value's are `helm`, `zarf` and `pack`, The default is 'helm'. `zarf` is allowed with `type="basic"`
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--credentials"></a>
### Nested Schema for `credentials`

Required:

- `credential_type` (String) The type of authentication used for accessing the registry. Supported values are 'secret', 'sts', 'basic', and 'noAuth'.

Optional:

- `access_key` (String) The access key for accessing the registry. Required if 'credential_type' is set to 'secret'.
- `arn` (String) The Amazon Resource Name (ARN) used for AWS-based authentication. Required if 'credential_type' is 'sts'.
- `external_id` (String) The external ID used for AWS STS (Security Token Service) authentication. Required if 'credential_type' is 'sts'.
- `password` (String, Sensitive) The password for basic authentication. Required if 'credential_type' is 'basic'.
- `secret_key` (String, Sensitive) The secret key for accessing the registry. Required if 'credential_type' is set to 'secret'.
- `tls_config` (Block List, Max: 1) TLS configuration for the registry. (see [below for nested schema](#nestedblock--credentials--tls_config))
- `username` (String) The username for basic authentication. Required if 'credential_type' is 'basic'.

<a id="nestedblock--credentials--tls_config"></a>
### Nested Schema for `credentials.tls_config`

Optional:

- `certificate` (String) Specifies the TLS certificate used for secure communication. Required for enabling SSL/TLS encryption.
- `insecure_skip_verify` (Boolean) Disables TLS certificate verification when set to true. Use with caution as it may expose connections to security risks.



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)