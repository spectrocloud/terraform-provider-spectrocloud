# Cluster templates tutorial example

End-to-end example accompanying the Spectro Cloud tutorial
[Standardize Clusters with Cluster Templates using Terraform](https://docs.spectrocloud.com/tutorials/clusters/cluster-templates/standardize-clusters-with-cluster-templates-terraform/).

This example shows how to standardize and later roll out an upgrade across a fleet of clusters using a
`spectrocloud_cluster_config_template`, rather than managing each cluster's profile version individually.
It provisions, per enabled cloud (AWS and/or Azure):
- A maintenance policy (`spectrocloud_cluster_config_policy`) defining a weekly upgrade window
- A cluster profile (v1.0.0: OS, Kubernetes, CNI, CSI, plus the Hello Universe sample app)
- A cluster template (`spectrocloud_cluster_config_template`) that pins a cluster profile version and
  attaches the maintenance policy
- Two clusters — `dev` and `prod` — both built from the same template, but each overriding the
  `app_replicas` profile variable to a different value (1 vs. 2)

## The Day-2 workflow this demonstrates

1. **Initial deploy:** with `create_new_profile_version` and `update_template_profile_version` both
   `false`, `terraform apply` creates the template pointing at profile v1.0.0, and both the `dev` and
   `prod` clusters.
2. **Add a new profile version:** set `create_new_profile_version = true` and re-apply. This creates
   profile v1.1.0, which adds the Kubecost add-on — the template still points at v1.0.0, so existing
   clusters are unaffected.
3. **Roll the template forward:** set `update_template_profile_version = true` and re-apply. The template
   now points at v1.1.0. This alone does not change already-running clusters — it changes what new
   clusters deployed from the template will use, and prepares the template for an explicit rollout.
4. **Trigger the upgrade:** set `upgrade_now_timestamp` to the current RFC3339 timestamp (for example
   `2026-05-12T14:30:00Z`) and re-apply. This immediately triggers an upgrade of every cluster launched
   from the template — both `dev` and `prod` — to the template's current profile version. Change this
   timestamp to any new value to re-trigger a later upgrade.

The maintenance policy's weekly schedule governs *routine* upgrade windows; `upgrade_now` is an explicit
override that upgrades immediately regardless of that schedule.

## Prerequisites

You will need the following before getting started:
1. A Palette API key, exported as the `SPECTROCLOUD_APIKEY` environment variable (or supplied via the
   `sc_api_key` variable).
2. A cloud account already registered in your Palette project settings for each cloud you enable.
3. For AWS: an existing EC2 key pair in the region you deploy to.
4. Terraform 1.9 or later.

## Instructions

Clone this repository to a local directory, and change directory to
`examples/tutorials/cluster-templates`. Proceed with the following:
1. From the current directory, copy the template variable file `terraform.template.tfvars` to a new
   file `terraform.tfvars`.
2. Set the `deploy-aws` and/or `deploy-azure` toggle(s), and fill in the placeholder (`REPLACE ME`)
   values relevant to those clouds.
3. Initialize and run terraform: `terraform init && terraform apply`. This deploys the maintenance
   policy, template, and both `dev`/`prod` clusters at profile v1.0.0.
4. Walk through the Day-2 workflow above by editing `terraform.tfvars` and re-applying at each step.

## Clean up

Run the destroy operation:

```shell
terraform destroy
```
