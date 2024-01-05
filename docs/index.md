---
page_title: "Spectro Cloud Provider"
subcategory: ""
description: |-
  The Spectro Cloud provider provides resources to interact with the Spectro Cloud management API (whether SaaS or on-prem).
---

# Spectro Cloud Provider

The Spectro Cloud provider provides resources to interact with the Spectro Cloud management API (whether SaaS or on-prem).

## What is Spectro Cloud?

The Spectro Cloud management platform brings the managed Kubernetes experience to users' own unique enterprise
Kubernetes infrastructure stacks running in any public cloud, or private cloud environments, allowing users to
not have to trade-off between flexibility and manageability. Spectro Cloud provides an as-a-service experience
to users by automating the lifecycle of multiple Kubernetes clusters based on user-defined Kubernetes
infrastructure stacks.

## Spectro Cloud account

This provider requires access to a valid Spectro Cloud account. Sign up for a free trial account [here](https://www.spectrocloud.com/free-trial/).
You may use your Spectro Cloud account credentials to access the Spectro Cloud management API or a Spectro Cloud API key. For more details on the authentication, navigate to the [authentication](#authentication) section.

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

->
Be sure to populate the `sc_host`, `sc_api_key`, and other terraform vars.

Copy one of the resource configuration files (e.g: spectrocloud_cluster_profile) from the _Resources_ documentation. Be sure to specify
all required parameters.

Next, run terraform using:

    terraform init && terraform apply

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

For questions or issues with the provider, please post your questions on the
provider GitHub [discussion board](https://github.com/spectrocloud/terraform-provider-spectrocloud/discussions).

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) The Spectro Cloud API key. Can also be set with the `SPECTROCLOUD_APIKEY` environment variable.
- `host` (String) The Spectro Cloud API host url. Can also be set with the `SPECTROCLOUD_HOST` environment variable. Defaults to https://api.spectrocloud.com
- `ignore_insecure_tls_error` (Boolean) Ignore insecure TLS errors for Spectro Cloud API endpoints. Defaults to false.
- `project_name` (String) The Palette project the provider will target. If no value is provided, the `Default` Palette project is used. The default value is `Default`.
- `retry_attempts` (Number) Number of retry attempts. Can also be set with the `SPECTROCLOUD_RETRY_ATTEMPTS` environment variable. Defaults to 10.
- `trace` (Boolean) Enable HTTP request tracing. Can also be set with the `SPECTROCLOUD_TRACE` environment variable. To enable Terraform debug logging, set `TF_LOG=DEBUG`. Visit the Terraform documentation to learn more about Terraform [debugging](https://developer.hashicorp.com/terraform/plugin/log/managing).
