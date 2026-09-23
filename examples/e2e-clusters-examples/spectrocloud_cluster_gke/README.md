# spectrocloud_cluster_gke - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_gke` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_gke`](../../resources/spectrocloud_cluster_gke).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_gcp` resource (the same
  cloud account type used by `spectrocloud_cluster_gcp`).
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using GCP Cloud Storage.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI + CSI + add-on pack
  stack, and a `profile_variables` block covering all 4 variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field - unlike the plain (non-GKE) `spectrocloud_cluster_gcp`,
    `cloud_config.project`/`region` **are** ForceNew here.
  - One `machine_pool` block, demonstrating this provider's simplest node pool schema: no
    `control_plane`/`control_plane_as_worker` fields (GKE's control plane is fully
    Google-managed), and no `min`/`max` autoscaling or `azs` placement, alongside taints,
    labels/annotations, `override_scaling`, and both override YAML passthroughs.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    the current `update_worker_pools_in_parallel` field (with its deprecated
    `update_worker_pool_in_parallel` alias shown commented), tags, description,
    `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, and local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. A GCP service account with permissions to create GKE clusters. See
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
