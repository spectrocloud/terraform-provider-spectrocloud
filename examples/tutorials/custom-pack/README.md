# Custom pack tutorial example

End-to-end example accompanying the Spectro Cloud tutorial
[Deploy a Custom Pack](https://docs.spectrocloud.com/tutorials/packs-registries/deploy-pack/).

This terraform configuration provisions a new AWS Kubernetes cluster whose cluster profile layers in
a custom add-on pack (the "Hello Universe" sample application from the tutorial) on top of the standard
OS, Kubernetes, CNI, and CSI layers. It will provision the following resources on Spectro Cloud:
- Cluster Profile (Ubuntu, Kubernetes, Calico CNI, AWS EBS CSI, plus your custom add-on pack)
- AWS Kubernetes Cluster

## Prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. An AWS cloud account already registered in your Palette project settings.
3. An AWS EC2 key pair created in the region where you will deploy the cluster.
4. The custom add-on pack itself, already created and published to a Palette pack registry (OCI or
   standard) by following the [Deploy a Custom Pack](https://docs.spectrocloud.com/tutorials/packs-registries/deploy-pack/)
   tutorial. This example deploys that pack — it does not create it.
5. Standard AWS credentials available via the environment (for example `AWS_ACCESS_KEY_ID` /
   `AWS_SECRET_ACCESS_KEY`), since this example also uses the `hashicorp/aws` provider to look up
   available Availability Zones.

## Instructions

Clone this repository to a local directory, and change directory to `examples/tutorials/custom-pack`.
Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Specify and update all the placeholder (`REPLACE ME`) values in the `terraform.tfvars` file.
3. Initialize and run terraform: `terraform init && terraform apply`.
4. Wait for the cluster creation to finish.

## Deploying to a different cloud

This example is written for AWS. To adapt it to Azure or GCP instead:
- Change the `cloud` argument on `spectrocloud_cluster_profile.profile` in `profile.tf`.
- Swap the OS, CNI, and CSI pack names/versions in `data.tf` for the equivalents on that cloud.
- Replace the `spectrocloud_cluster_aws` resource in `cluster.tf` with `spectrocloud_cluster_azure` or
  `spectrocloud_cluster_gcp`, and update `cloud_config` to match.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
