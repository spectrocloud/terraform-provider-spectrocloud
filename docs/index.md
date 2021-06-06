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
  username     = var.sc_username     # Username of the user (or specify with SPECTROCLOUD_USERNAME env var)
  password     = var.sc_password     # Password (or specify with SPECTROCLOUD_PASSWORD env var)
  project_name = var.sc_project_name # Project name (e.g: Default)
}
```

Create or append to a `terraform.tfvars` file:

```terraform
# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud API endpoint}" #e.g: api.spectrocloud.com (for SaaS)
sc_username     = "{enter Spectro Cloud username}"     #e.g: user1@abc.com
sc_password     = "{enter Spectro Cloud password}"     #e.g: supereSecure1!
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default
```

->
Be sure to populate the `username`, `password`, and other terraform vars.

Copy one of the resource configuration files (e.g: spectrocloud_cluster_profile) from the _Resources_ documentation. Be sure to specify
all required parameters.

Next, run terraform using:

    terraform init && terraform apply

Detailed schema definitions for each resource are listed in the _Resources_ menu on the left.

For an end-to-end example of provisioning Spectro Cloud resources, visit:
[Spectro Cloud E2E Examples](https://github.com/spectrocloud/terraform-provider-spectrocloud/tree/main/examples/e2e).

## Support

For questions or issues with the provider, please post your questions on the
provider GitHub [discussion board](https://github.com/spectrocloud/terraform-provider-spectrocloud/discussions).

## Schema

### Required

- **password** (String, Sensitive)
- **username** (String)

### Optional

- **host** (String)
- **ignore_insecure_tls_error** (Boolean)
- **project_name** (String)
