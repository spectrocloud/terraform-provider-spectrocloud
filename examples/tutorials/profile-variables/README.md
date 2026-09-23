# Cluster profile variables tutorial example

End-to-end example accompanying the Spectro Cloud tutorial
[Deploy Cluster with Profile Variables](https://docs.spectrocloud.com/tutorials/profiles/cluster-profile-variables/).

This example deploys a WordPress application on a Kubernetes cluster (AWS, Azure, and/or GCP), and
demonstrates the difference between a cluster profile with no variables and one using
[profile variables](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs/resources/cluster_profile#profile_variables)
to make Day-2 configuration changes without editing the manifest itself:
- **Profile version `1.0.0`** deploys WordPress with fixed values baked into `manifests/wordpress-default.yaml`
  (namespace `wordpress`, 1 replica, port 80).
- **Profile version `1.1.0`** deploys the same chart via `manifests/wordpress-variables.yaml`, which
  references `{{ .spectro.var.wordpress_namespace }}`, `{{ .spectro.var.wordpress_replica }}`, and
  `{{ .spectro.var.wordpress_port }}` instead — with a `profile_variables` block on the profile resource
  supplying those values, so they can be changed later via Palette without editing the manifest.

## Choosing which cloud(s) to deploy to, and which profile version

| Variable | Provider | Description | Default |
|---|---|---|---|
| `deploy-aws` | AWS | Enable to deploy a cluster to AWS. | `false` |
| `deploy-azure` | Azure | Enable to deploy a cluster to Azure. | `false` |
| `deploy-gcp` | GCP | Enable to deploy a cluster to GCP. | `false` |

Each cloud also has a `deploy-<cloud>-var` toggle: when `false` (default), the cluster deploys profile
version `1.0.0` (no variables); when `true`, it deploys version `1.1.0` (with `profile_variables`), reading
its values from `wordpress_replica`, `wordpress_namespace`, and `wordpress_port`.

## Prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. A cloud account already registered in your Palette project settings for each cloud you enable.
3. For AWS: an existing EC2 key pair in the region you deploy to.
4. Terraform 1.9 or later.

## Instructions

Clone this repository to a local directory, and change directory to
`examples/tutorials/profile-variables`. Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Set the `deploy-<cloud>` toggle(s) for the cloud(s) you want, and fill in all the placeholder
   (`REPLACE ME`) values relevant to those clouds.
3. Initialize and run terraform: `terraform init && terraform apply`.
4. Wait for the cluster creation to finish.
5. To see profile variables in action, set the matching `deploy-<cloud>-var` to `true`, fill in
   `wordpress_replica`, `wordpress_namespace`, and `wordpress_port`, and re-apply. The cluster switches
   from profile version `1.0.0` to `1.1.0` in place.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
