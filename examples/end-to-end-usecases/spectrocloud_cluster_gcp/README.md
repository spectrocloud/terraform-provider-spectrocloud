# spectrocloud_cluster_gcp - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_gcp` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_gcp`](../../resources/spectrocloud_cluster_gcp).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_gcp` resource using a
  service account JSON key.
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using the `gcp_storage_config` block (GCP Cloud Storage), matching this cluster's
  cloud provider.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI + CSI + add-on pack
  stack, and a `profile_variables` block covering all 4 variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field. Notably, `cloud_config` is **not** ForceNew for this resource -
    unusual compared to most other cluster resources in this provider, where the whole block is
    typically ForceNew.
  - Two `machine_pool` blocks: a 3-node control plane and a worker pool, both with required
    `azs` placement (unlike some clouds where it's optional), plus taints, labels/annotations,
    `override_scaling`, `override_kubeadm_configuration`, and `override_cluster_api_config`. Note:
    this machine_pool has no `min`/`max` autoscaling fields.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    `update_worker_pools_in_parallel`, tags, description, `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, and local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. A GCP service account with permissions to create the resources Palette provisions. See
   [GCP Service Account](https://docs.spectrocloud.com/clusters/?clusterType=google_cloud_cluster#creatingagcpcloudaccount).
2. A second (or the same) GCP service account/project to use as the backup storage location.

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
