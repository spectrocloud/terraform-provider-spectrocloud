![release](https://github.com/spectrocloud/terraform-provider-spectrocloud/workflows/release/badge.svg)

# Terraform Provider spectrocloud

Terraform Provider for Spectro Cloud.

## Pre-Requisites

To use this Spectro Cloud provider, you must meet the following requirements:
- Spectro Cloud account ([Sign-up for a free trial account](https://www.spectrocloud.com/free-trial/) )
- Terraform (minimum version 0.13+)
- Kubernetes/Kubectl CLI (minimum version 1.16+)

## Usage

For an end end-to-end cluster provisioning example, please follow the appropriate guide under
[Spectro Cloud E2E Examples](examples/e2e/).

Examples of other managed resources are also available in the [examples/resources/](examples/resources/) directory.

Detailed documentation on supported data sources and resources are available on the
[Terraform Spectro Cloud Provider Documentation](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs).

## Develop

- Hack away
- Make sure to run `go generate` after your final commit
- Send in a PR


### Documentation

The documentation for each respective resource is found in the [docs](/docs) folder. Please ensure you are following the Terraform Registry [documentation guidance](https://developer.hashicorp.com/terraform/registry/providers/docs). To preview documentation changes, please utilize the [Terraform Registry Preview Tool](https://registry.terraform.io/tools/doc-preview).

## Support

For questions or issues with the provider, please post your questions on the provider [discussion board](/discussions).
