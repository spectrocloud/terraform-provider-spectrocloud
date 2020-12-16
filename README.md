# Terraform Provider spectrocloud

This repo is a companion repo to the [Call APIs with Terraform Providers](https://learn.hashicorp.com/collections/terraform/providers) Learn collection. 

In the collection, you will use the spectrocloud provider as a bridge between Terraform and the Spectro Cloud API. Then, extend Terraform by recreating the Spectro Cloud provider. By the end of this collection, you will be able to take these intuitions to create your own custom Terraform provider. 

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-spectrocloud
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```
