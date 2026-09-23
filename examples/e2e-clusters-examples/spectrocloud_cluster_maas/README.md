# spectrocloud_cluster_maas - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_maas` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_maas`](../../resources/spectrocloud_cluster_maas).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_maas` resource, routed
  through an existing Private Cloud Gateway (PCG).
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio) - the common choice for
  on-prem clusters, alongside the AWS/GCP/Azure alternatives shown commented.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI + CSI + add-on pack
  stack, and a `profile_variables` block covering all 4 variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field: `domain`, `ssh_keys`, `enable_lxd_vm`, `ntp_servers`, and the
    `override_cluster_api_config` YAML passthrough for CAPMAAS properties.
  - Two `machine_pool` blocks: a control-plane pool (`control_plane`/`control_plane_as_worker`)
    and an autoscaling worker pool, both using MAAS's `instance_type` (`min_memory_mb`/`min_cpu`
    resource requirements instead of a fixed VM size), `placement.resource_pool`, `azs`, and
    `node_tags` for MAAS automatic-tag-based node placement, plus taints, labels/annotations,
    `override_scaling`, and both override YAML passthroughs.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    `update_worker_pools_in_parallel`, tags, description, `cluster_meta_attribute`.
  - The optional `hyper_shift_config` block (for HyperShift/OpenShift hosted control planes on
    MAAS) is shown commented, since this example provisions a standard control plane instead.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, `location_config`, and
  local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. A Private Cloud Gateway (PCG) already installed and reachable from your MAAS environment. See
   [Private Cloud Gateway](https://docs.spectrocloud.com/clusters/data-center/deployment-installation/).
2. A MAAS API endpoint and API key with permissions to provision machines.
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
