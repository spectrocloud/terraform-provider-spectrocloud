---
page_title: "Spectro Cloud Provider"
subcategory: ""
description: |-
  The Spectro Cloud provider provides resources to interact with the Spectro Cloud management API (whether SaaS or on-prem).
---

# Spectro Cloud Provider

The Spectro Cloud provider provides resources to interact with Palette and Palette VerteX through Infrastructure as code. The provider supports both SaaS and on-prem deployments of Palette and Palette VerteX.

## What is Palette?

Palette brings the managed Kubernetes experience to users' own unique enterprise
Kubernetes infrastructure stacks deployed in any public cloud, or private cloud environments. Palette allows users to
not have to trade-off between flexibility and manageability. Palette provides a platform-as-a-service experience
to users by automating the lifecycle of multiple Kubernetes clusters based on user-defined Kubernetes
infrastructure stacks.

## Palette Account

To get started with Palette, sign up for an account [here](https://www.spectrocloud.com/get-started).
Use your Palette [API key](https://docs.spectrocloud.com/user-management/authentication/api-key/create-api-key) to authenticate. For more details on the authentication, navigate to the [authentication](#authentication) section.

## Example Usage

Create a `providers.tf` file with the following:

```terraform
terraform {
  required_providers {
    spectrocloud = {
      version = ">= 0.1"
      source  = "spectrocloud/spectrocloud"
    }
  }
}

provider "spectrocloud" {
  host         = var.sc_host         # Spectro Cloud endpoint (defaults to api.spectrocloud.com)
  api_key      = var.sc_api_key      # API key (or specify with SPECTROCLOUD_APIKEY env var)
  project_name = var.sc_project_name # Project name (e.g: Default)
}
```

Copy `terraform.template.tfvars` file to a `terraform.tfvars` file and modify its content:

```terraform
##################################################################################
# Spectro Cloud credentials
##################################################################################
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default
```

Be sure to populate the `sc_host`, `sc_api_key`, and other terraform vars.

Copy one of the resource configuration files (e.g: spectrocloud_cluster_profile) from the _Resources_ documentation. Be sure to specify
all required parameters.

Next, run terraform using:

```console
terraform init && terraform apply
```

Detailed schema definitions for each resource are listed in the _Resources_ menu on the left.

For an end-to-end example of provisioning Spectro Cloud resources, visit:
[Spectro Cloud E2E Examples](https://github.com/spectrocloud/terraform-provider-spectrocloud/tree/main/examples/e2e).

## Environment Variables

Credentials and other configurations can be provided through environment variables. The following environment variables are availabe.

- `SPECTROCLOUD_HOST`
- `SPECTROCLOUD_APIKEY`
- `SPECTROCLOUD_TRACE`
- `SPECTROCLOUD_RETRY_ATTEMPTS`


## Authentication
You can use an API key to authenticate with Spectro Cloud. Visit the User Management API Key [documentation](https://docs.spectrocloud.com/user-management/user-authentication/#usingapikey) to learn more about Spectro Cloud API keys.
```shell
export SPECTROCLOUD_APIKEY=5b7aad.........
```

```hcl
provider "spectrocloud" {}
```

## Feature Flags

The provider accepts optional feature flags through the `feature_flag` map argument in the provider block. Unknown keys are ignored.

### disable_addon_deployment_resource

When set to `true`, the provider disallows the [`spectrocloud_addon_deployment`](resources/addon_deployment.md) resource. Any configuration that references this resource fails during `terraform plan` (including refresh) and apply with an error indicating the feature flag is disabled. Defaults to `false`.

`cluster_profile` and `cluster_template` are **mutually exclusive** on `spectrocloud_cluster_*` resources. The provider returns an error if both are set in the same resource.

When this flag is enabled:

- **`cluster_profile` only** — On read, the provider refreshes every profile attached to the cluster from the Palette API into the top-level `cluster_profile` block (including addon profiles previously managed by `spectrocloud_addon_deployment`). Pack and variable fields in state follow what you declare in configuration.
- **`cluster_template` only** — The provider does **not** modify top-level `cluster_profile`. Profiles are managed inside `cluster_template` (nested `cluster_profile` blocks). Read continues to refresh `cluster_template` variables from the API via the existing template flow.
- **Neither** — No profile sync from this flag.

To destroy existing addon deployments still in Terraform state, set the flag back to `false`, run `terraform destroy` on those resources, then set the flag to `true` again if you want to keep `spectrocloud_addon_deployment` blocked.

```terraform
provider "spectrocloud" {
  host    = var.sc_host
  api_key = var.sc_api_key

  feature_flag = {
    disable_addon_deployment_resource = true
  }
}
```

## Feature Preview

The provider accepts optional feature preview flags through the `feature_preview` map argument in the provider block.

### immutable-clusterprofiles

When set to `true`, `spectrocloud_cluster_profile` uses the standard Terraform Plugin SDK v2 immutable-versioned-resource pattern. Version bumps trigger a Terraform replacement (`ForceNew`) instead of an in-place update. Combined with `skip_destroy = true` and `lifecycle { create_before_destroy = true }` on the resource, each version is preserved in Palette while Terraform state advances to the new version. See [`spectrocloud_cluster_profile`](resources/cluster_profile.md) for full usage details.

```terraform
provider "spectrocloud" {
  host    = var.sc_host
  api_key = var.sc_api_key

  feature_preview = {
    "immutable-clusterprofiles" = true
  }
}
```

## Import
Starting with Terraform v1.5.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import resources into your state file.

Each resource type has its own specific requirements for the import process. We recommend you refer to the documentation for each resource to better understand the exact format of the `id` and any other required parameters.
The `import` block specifies the resource you want to import and its unique identifier with the following structure:

```terraform
import {
  to = <resource>.<name>
  id = "<unique_identifier>"
}
```

- `<resource>`: The type of the resource you are importing.
- `<name>`: A name you assign to the resource within your Terraform configuration.
- `<unique_identifier>`: The ID of the resource you are importing. This can include additional context if required.

### Examples

The following examples showcase how to import a resource. Some resource requires the context to be specified during the import action. The context refers to the Palette scope. Allowed values are either `project` or `tenant`. 

####  Import With Context

When importing resources that require additional context, the `id` is followed by a context, separated by a colon.

   ```terraform
   import {
     to = spectrocloud_cluster_aks.example
     id = "example_id:project"
   }
   ```

  You can also import a resource using the Terraform CLI and the `import` command.

   ```console
   terraform import spectrocloud_cluster_aks.example example_id:project
   ```

    Specify' tenant' after the colon if you want to import a resource at the tenant scope. 

  ```terraform
  import {
    to = spectrocloud_cluster_aks.example
    id = "example_id:tenant"
  }
  ```

  Example of importing a resource with the tenant context through the Terraform CLI.

  ```console
  terraform import spectrocloud_cluster_aks.example example_id:tenant
  ```

~> Ensure you have tenant admin access when importing a resource at the tenant scope.

#### Import Without Context

For resources that do not require additional context, the `id` is the only provided argument. The following is an example of a resource that does not require the context and only provides the ID.

   ```terraform
   import {
     to = spectrocloud_cluster_profile.example
     id = "id"
   }
   ```

   Below is an example of using the Terraform CLI and the `import` command without specifying the context.

   ```console
   terraform import spectrocloud_cluster_profile.example id
   ```


## Support

For questions or issues with the provider, open up an issue in the provider GitHub [discussion board](https://github.com/spectrocloud/terraform-provider-spectrocloud/discussions).

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) The Spectro Cloud API key. Can also be set with the `SPECTROCLOUD_APIKEY` environment variable.
- `feature_flag` (Map of Boolean) Optional provider feature flags (map of booleans). Unknown keys are ignored. Set `disable_addon_deployment_resource` to `true` to block the `spectrocloud_addon_deployment` resource during plan and apply.
- `feature_preview` (Map of Boolean) A map of feature preview flags. Supported flags: `immutable-clusterprofiles`. 

The `immutable-clusterprofiles` flag enables the standard Terraform Plugin SDK v2 immutable-versioned-resource pattern for `spectrocloud_cluster_profile`. When set, the resource's `version` field becomes `ForceNew` (changes trigger a Terraform replacement instead of an in-place update), the new `skip_destroy` schema field is honored, and the Create function clones from any existing version of the lineage to produce the new immutable version. The Terraform resource id is set once at Create time and never mutates mid-update, so it respects the SDK v2 contract that a resource's primary id is stable across in-place updates. 

Without the flag, `spectrocloud_cluster_profile` uses its legacy in-place mutation behavior (PUT-based updates that overwrite the previous version). The flag is purely opt-in; existing user configurations are unaffected.
- `host` (String) The Spectro Cloud API host url. Can also be set with the `SPECTROCLOUD_HOST` environment variable. Defaults to https://api.spectrocloud.com
- `ignore_insecure_tls_error` (Boolean) Ignore insecure TLS errors for Spectro Cloud API endpoints. ⚠️ WARNING: Setting this to true disables SSL certificate verification and makes connections vulnerable to man-in-the-middle attacks. Only use this in development/testing environments or when connecting to self-signed certificates in trusted networks. Defaults to false.
- `project_name` (String) The Palette project the provider will target. If no value is provided, the `Default` Palette project is used. The default value is `Default`.
- `retry_attempts` (Number) Number of retry attempts. Can also be set with the `SPECTROCLOUD_RETRY_ATTEMPTS` environment variable. Defaults to 10.
- `trace` (Boolean) Enable HTTP request tracing. Can also be set with the `SPECTROCLOUD_TRACE` environment variable. To enable Terraform debug logging, set `TF_LOG=DEBUG`. Visit the Terraform documentation to learn more about Terraform [debugging](https://developer.hashicorp.com/terraform/plugin/log/managing).
