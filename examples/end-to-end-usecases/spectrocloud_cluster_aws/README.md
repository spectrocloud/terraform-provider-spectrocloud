# spectrocloud_cluster_aws - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_aws` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_aws`](../../resources/spectrocloud_cluster_aws).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_aws` resource, using
  static access/secret key credentials (the "secret" auth type); the "sts" (assumable role) and
  "pod-identity" (EKS Pod Identity) alternatives are shown commented.
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using AWS S3 with static credentials.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI + CSI + add-on pack
  stack, `context`, and a `profile_variables` block covering all 4 variable formats (hidden
  string, version with a regex constraint, dropdown, multiline).
- **Cluster** (`cluster.tf`):
  - `tags_map`, `cluster_type` (set once at creation), and every `cloud_config` field.
  - Two `machine_pool` blocks: a 3-node control plane with dynamic AZ placement, and a worker
    pool demonstrating spot capacity (`capacity_type`/`max_price`), autoscaling (`min`/`max` +
    `override_scaling`), taints, additional labels/annotations/security groups,
    `node_repave_interval`, `skip_k8s_upgrade`, and both `override_kubeadm_configuration` and
    `override_cluster_api_config` YAML passthroughs. Static AZ/subnet placement
    (`az_subnets`) and AWS Dedicated Host placement (`host_resource_group_arn`) are shown as
    commented, mutually-exclusive alternatives.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy` (pointing at the backup storage location created above), `scan_policy`, two
    `cluster_rbac_binding` blocks, `namespaces`, and `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    `update_worker_pools_in_parallel`, tags, description, `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): uses the `hashicorp/local` provider to write
  both the regular and cluster-admin kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfig/admin kubeconfig, `cloud_config_id`, and the
  local kubeconfig file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. An AWS account with permissions to create the resources Palette provisions. See
   [AWS Cloud Account](https://docs.spectrocloud.com/clusters/?clusterType=aws_cluster#awscloudaccountpermissions).
2. An existing EC2 key pair in the target region.
3. An S3 bucket to use as the backup storage location.

## Usage

1. Copy `terraform.template.tfvars` to `terraform.tfvars` and fill in every placeholder value.
2. `terraform init && terraform apply`.
3. `kubeconfig.tf` writes the kubeconfig files to the current directory automatically:
   ```shell
   export KUBECONFIG=$(terraform output -raw kubeconfig_path)
   kubectl get pod -A
   ```

## Clean up

```shell
terraform destroy
```
