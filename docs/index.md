---
page_title: "Spectro Cloud Provider"
subcategory: ""
description: |-
  The Spectro Cloud provider provides resources to interact with Spectro Cloud management API (SaaS or on-prem).
---

# Spectro Cloud Provider

The Spectro Cloud provider provides resources to interact with Spectro Cloud management API (SaaS or on-prem).

## What is Spectro Cloud?

The Spectro Cloud management platform brings the managed Kubernetes experience to users' own unique enterprise
Kubernetes infrastructure stacks running in any public cloud, or private cloud environments, allowing users to
not have to trade-off between flexibility and manageability. Spectro Cloud provides an as-a-service experience
to users by automating the lifecycle of multiple Kubernetes clusters based on user-defined Kubernetes
infrastructure stacks.

## Spectro Cloud account

This provider requires access to a valid Spectro Cloud account.

If you haven't already, please signup for a free Spectro Cloud account here: [Spectro Cloud Signup](https://www.spectrocloud.com/free-trial/).

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
# Credentials
host         = "console.spectrocloud.com"
username     = "user1@abc.com" # Username of the user (or specify with SPECTROCLOUD_USERNAME env var)
password     = "superSecure1!" # Password of the user (or specify with SPECTROCLOUD_PASSWORD env var)
project_name = "Default"       # Project name (e.g: Default)
```

->
Be sure to populate the `username`, `password`, and other terraform vars.

Ok

-> Be sure to populate the `username`, `password`, and other terraform vars.

Next, run terraform using:

    terraform init && terraform apply

Detailed schema definitions for each resource are listed in the _Resources_ menu on the left.

For an end-to-end example of provisioning Spectro Cloud resources, visit: [...](https://github.com).

## Support

For questions or issues with the provider, please post your questions on the
provider GitHub [discussion board](https://github.com/spectrocloud/terraform-provider-spectrocloud/discussions).

## Schema

### Required

- **password** (String, Sensitive)
- **username** (String)

### Optional

- **host** (String)
- **project_name** (String)
