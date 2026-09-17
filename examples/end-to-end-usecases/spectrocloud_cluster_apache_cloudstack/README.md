# spectrocloud_cluster_apache_cloudstack - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_apache_cloudstack` and its supporting resources in one working,
self-contained configuration - it creates its own cloud account and backup storage location
rather than looking up existing ones. For the minimal happy-path version of just the cluster
resource, see
[`examples/resources/spectrocloud_cluster_apache_cloudstack`](../../resources/spectrocloud_cluster_apache_cloudstack).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_apache_cloudstack`
  resource, routed through an existing Private Cloud Gateway (PCG).
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio) - the common choice for
  on-prem clusters, alongside the AWS/GCP/Azure alternatives shown commented.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI + CSI + add-on pack
  stack, and a `profile_variables` block covering all 4 variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field: `ssh_key_name`, `control_plane_endpoint`, `sync_with_cks`,
    `project`, and a `zone` block with a nested `network` (and, commented, a `vpc` block for
    VPC-based deployments). Unlike most other cloud clusters, this entire block is **not**
    ForceNew - it updates in place.
  - Two `machine_pool` blocks with a required `offering` (CloudStack compute size), an optional
    per-pool `network`/`template` override, autoscaling (`min`/`max`), taints,
    labels/annotations, `override_scaling`, and both override YAML passthroughs.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`,
    `host_config`, and a `timeouts` block for per-operation timeout overrides.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    `update_worker_pools_in_parallel`, tags, description, `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, `location_config`, and
  local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. A Private Cloud Gateway (PCG) already installed and reachable from your CloudStack
   environment.
2. A CloudStack API endpoint, API key, and secret key with permissions to provision instances.
3. A Minio (or other S3-compatible) endpoint to use as the backup storage location.

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
