![release](https://github.com/spectrocloud/terraform-provider-spectrocloud/workflows/release/badge.svg)

# Terraform Provider spectrocloud

Terraform Provider for Spectro Cloud.

## Pre-Requisites

To use this Spectro Cloud provider, you must meet the following requirements:
- Spectro Cloud account ([Sign-up for a free account](https://www.spectrocloud.com/free-trial/) )
- Terraform 0.13+ (e.g: `brew install terraform`)

## Usage

For an end end-to-end cluster provisioning example, please follow the appropriate guide under
[Spectro Cloud E2E Examples](examples/e2e/).

Examples of other managed resources are also available in the [examples/resources/](examples/resources/) directory.

Detailed documentation on supported data sources and resources are available on the
[Terraform Spectro Cloud Provider Documentation](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs).

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-spectrocloud
```
