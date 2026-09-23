---
page_title: "spectrocloud_backup_storage_location Resource - terraform-provider-spectrocloud"
subcategory: "Backup Location"
description: |-
  Resource for managing backup storage locations in Spectro Cloud.
---

# spectrocloud_backup_storage_location (Resource)

  Resource for managing backup storage locations in Spectro Cloud.

## Example Usage

```terraform
# Common attributes across all backup storage locations, regardless of provider:
#   name             - Required.
#   storage_provider - Optional, ForceNew, default "aws". Allowed: "aws", "minio", "gcp", "azure".
#                       This is the only ForceNew attribute on the resource - changing it
#                       recreates the backup location under a new provider.
#   context          - Optional, default "project". Allowed: "project", "tenant".
#   is_default       - Optional, default false. Whether new backups use this location by default.
#   region           - Optional in the schema, but a CustomizeDiff (schemaValidationForLocationProvider)
#                       requires it (along with `s3`/`bucket_name`) whenever storage_provider is
#                       "aws" or "minio", and forbids it entirely for "azure"/"gcp".
#   bucket_name      - Same CustomizeDiff: required for "aws"/"minio" (with `s3`/`region`) and for
#                       "gcp" (with `gcp_storage_config`); forbidden for "azure".
#   ca_cert          - Same CustomizeDiff: only meaningful for "aws"/"minio"; forbidden for
#                       "azure"/"gcp".
# The same CustomizeDiff also forbids `azure_storage_config`/`gcp_storage_config` when
# storage_provider is "aws"/"minio", and forbids `s3` when it's "azure"/"gcp" - each provider's
# block below is mutually exclusive with the others in practice, not just by convention.
# All of the above (aside from storage_provider) update in place.
#
# s3 block (used by bsl_s3/bsl_minio/bsl_sts below - one block shape shared by AWS and any
# S3-compatible service, at most one per resource):
#   credential_type     - Required. Allowed: "secret" (static access/secret key) or "sts"
#                         (assumable IAM role).
#   access_key          - Required when credential_type = "secret", sensitive.
#   secret_key          - Required when credential_type = "secret", sensitive.
#   arn                 - Required when credential_type = "sts". The IAM role ARN to assume.
#   external_id         - Optional. Used alongside `arn` for cross-account STS role assumption
#                         when credential_type = "sts".
#   s3_force_path_style - Optional. Forces path-style S3 URLs (bucket in the path, not the
#                         hostname) - needed for most S3-compatible services (e.g. Minio).
#   s3_url              - Optional. The S3 endpoint URL.

// S3 backup location, using static access/secret key credentials
resource "spectrocloud_backup_storage_location" "bsl_s3" {
  name             = "project-dev-bsl-s3"
  context          = "project"
  storage_provider = "aws"
  is_default       = false
  region           = "us-east-1"
  bucket_name      = "project-backup-bucket-s3"
  # ca_cert          = "REPLACE_ME"

  # Static access/secret key credentials - see the s3 block reference in the top comment.
  s3 {
    credential_type     = "secret"
    access_key          = var.aws_access_key
    secret_key          = var.aws_secret_key
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-1.amazonaws.com"
  }
}

// Minio (S3-compatible) backup location, using static credentials
resource "spectrocloud_backup_storage_location" "bsl_minio" {
  name             = "project-dev-minio-bsl"
  context          = "project"
  storage_provider = "minio"
  is_default       = false
  region           = "us-east-2"
  bucket_name      = "project-backup-bucket-minio"

  # Same s3 block shape as bsl_s3 above; Minio (and most other S3-compatible services) need
  # s3_force_path_style = true, unlike AWS's own endpoint.
  s3 {
    credential_type     = "secret"
    access_key          = var.aws_access_key
    secret_key          = var.aws_secret_key
    s3_force_path_style = true
    s3_url              = "http://10.90.78.23"
  }
}

// GCP backup location
resource "spectrocloud_backup_storage_location" "bsl_gcp" {
  name             = "project-dev-gcp"
  context          = "project"
  storage_provider = "gcp"
  is_default       = false
  bucket_name      = "project-backup-bucket-gcp"

  # gcp_storage_config block (Required when storage_provider = "gcp", at most one):
  #   project_id           - Required.
  #   gcp_json_credentials - Required, sensitive.
  gcp_storage_config {
    project_id           = "test-gcp-project-id"
    gcp_json_credentials = var.gcp_json_credentials
  }
}

// Azure backup location
resource "spectrocloud_backup_storage_location" "bsl_azure" {
  name             = "project-dev-azure-bsl"
  context          = "project"
  storage_provider = "azure"
  is_default       = false

  # azure_storage_config block (Required when storage_provider = "azure", at most one - every
  # field here is Required):
  #   stock_keeping_unit  - Azure storage account SKU, e.g. "Standard_LRS".
  #   azure_client_secret - Sensitive.
  #   (container_name, storage_name, resource_group, azure_tenant_id, azure_client_id,
  #   subscription_id carry no further constraints beyond being Required.)
  azure_storage_config {
    container_name      = "test-container"
    storage_name        = "test-storage"
    stock_keeping_unit  = "Standard_LRS"
    resource_group      = "test-resource-group"
    azure_tenant_id     = "test-azure-tenant-id"
    azure_client_id     = "test-azure-client-id"
    subscription_id     = "test-azure-subscription-id"
    azure_client_secret = var.azure_client_secret
  }
}

// S3 backup location, using an assumable IAM role (STS) instead of static keys
resource "spectrocloud_backup_storage_location" "bsl_sts" {
  name        = "tenant-dev-1"
  context     = "tenant"
  is_default  = false
  region      = "us-east-2"
  bucket_name = "tenant-backup-bucket-sts"

  # Same s3 block shape as bsl_s3 above, but credential_type = "sts" - see the top comment for
  # arn/external_id.
  s3 {
    credential_type     = "sts"
    arn                 = var.aws_sts_role_arn
    external_id         = var.aws_external_id
    s3_force_path_style = false
    s3_url              = "https://s3.us-east-2.amazonaws.com"
  }
}
```

## Import

Backup Storage Locations can be imported using either a simple ID format or with explicit context specification. This resource supports both project and tenant contexts.

### Simple Import (defaults to project context)

```bash
terraform import spectrocloud_backup_storage_location.example {bsl_id}/{bsl_name}:project
```

### Context-specific Import

```bash
terraform import spectrocloud_backup_storage_location.example <bsl_id>/<bsl_name>: <project>/<tenant>
```

Where:
- `<bsl_id>` is the Backup Storage Location ID
- `project` or `tenant` specifies the context where the backup storage location exists

**Import behavior:**
- If no context is specified, it defaults to `project` context
- If the resource is not found in the specified context, the import will automatically try the other context
- The import will automatically populate all configuration fields from the Spectro Cloud API, including the correct context, storage provider, and all provider-specific settings

After import, you can run `terraform plan` to see the current configuration and make any necessary adjustments.


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the backup storage location. This is a unique identifier for the backup location.

### Optional

- `azure_storage_config` (Block List, Max: 1) Azure storage settings for configuring the backup storage location. (see [below for nested schema](#nestedblock--azure_storage_config))
- `bucket_name` (String) The name of the storage bucket where backups are stored. This is relevant for S3 or S3-compatible(minio) or gcp storage services.
- `ca_cert` (String) An optional CA certificate used for SSL connections to ensure secure communication with the storage provider. This is relevant for S3 or S3-compatible(minio) storage services.
- `context` (String) The context of the backup storage location. Allowed values are `project` or `tenant`. Default value is `project`. If  the `project` context is specified, the project name will sourced from the provider configuration parameter [`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema).
- `gcp_storage_config` (Block List, Max: 1) GCP storage settings for configuring the backup storage location. (see [below for nested schema](#nestedblock--gcp_storage_config))
- `is_default` (Boolean) Specifies if this backup storage location should be used as the default location for storing backups.
- `region` (String) The region where the backup storage is located, typically corresponding to the region of the cloud provider. This is relevant for S3 or S3-compatible(minio) storage services.
- `s3` (Block List, Max: 1) S3-specific settings for configuring the backup storage location. (see [below for nested schema](#nestedblock--s3))
- `storage_provider` (String) The storage location provider for backup storage. Allowed values are `aws` or `minio` or `gcp` or `azure`. Default value is `aws`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--azure_storage_config"></a>
### Nested Schema for `azure_storage_config`

Required:

- `azure_client_id` (String) Unique client Id from Azure console.
- `azure_client_secret` (String, Sensitive) Azure secret for authentication.
- `azure_tenant_id` (String) Unique tenant Id from Azure console.
- `container_name` (String) The container name.
- `resource_group` (String) The resource group name.
- `stock_keeping_unit` (String) The stop-keeping unit. eg: `Standard_LRS`
- `storage_name` (String) The storage name.
- `subscription_id` (String) Unique subscription Id from Azure console.


<a id="nestedblock--gcp_storage_config"></a>
### Nested Schema for `gcp_storage_config`

Required:

- `gcp_json_credentials` (String, Sensitive) The GCP credentials in JSON format. These credentials are required to authenticate and manage.
- `project_id` (String) The GCP project ID.


<a id="nestedblock--s3"></a>
### Nested Schema for `s3`

Required:

- `credential_type` (String) The type of credentials used to access the S3 storage. Supported values are 'secret' for static credentials and 'sts' for temporary, token-based credentials.

Optional:

- `access_key` (String, Sensitive) The access key for S3 authentication (credential), required if 'credential_type' is set to 'secret'.
- `arn` (String) The Amazon Resource Name (ARN) of the IAM role to assume for accessing S3 when using 'sts' credentials.
- `external_id` (String) An external ID used for cross-account access to the S3 storage when using 'sts' credentials.
- `s3_force_path_style` (Boolean) A boolean flag indicating whether to enforce the path-style URL for accessing S3.
- `s3_url` (String) The S3 URL endpoint.
- `secret_key` (String, Sensitive) The secret key for S3 authentication, required if 'credential_type' is set to 'secret'.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)