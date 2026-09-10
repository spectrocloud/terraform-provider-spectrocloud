# spectrocloud_cluster_custom_cloud - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_custom_cloud` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_custom_cloud`](../../resources/spectrocloud_cluster_custom_cloud).

`spectrocloud_cluster_custom_cloud` is a generic escape hatch for cloud providers that don't have
a first-class `spectrocloud_cluster_<cloud>` resource: instead of structured attributes like
`cloud_config.datacenter` or `machine_pool.instance_type`, the cluster is described almost
entirely as raw Cluster API YAML. This example uses **Nutanix** - the reference custom cloud
provider used throughout this provider's own test suite - but the same pattern applies to any
custom cloud provider registered in Palette.

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_custom` resource, whose
  generic `credentials` map holds Nutanix Prism Central credentials, routed through an existing
  Private Cloud Gateway (PCG).
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio) - the common choice for
  on-prem infrastructure like Nutanix, alongside the AWS/GCP/Azure alternatives shown commented.
- **Cluster profile** (`clusterprofile.tf`): Kubernetes + CNI + one add-on pack, and a
  `profile_variables` block covering all 4 variable formats. Custom cloud profiles skip the OS
  and CSI layers other cloud types need - node provisioning is described directly in the raw
  Cluster API YAML instead.
- **Cluster** (`cluster.tf`):
  - `cloud_config.values`: the cluster-scoped Cluster API manifests (Secret, ConfigMap,
    NutanixCluster, MachineHealthCheck), rendered via `templatefile()` from
    [`config_templates/cloud_config.yaml`](config_templates/cloud_config.yaml).
  - `cloud_config.overrides` / `machine_pool.overrides`: the schema's YAML-patching mechanism -
    template variables, wildcard field matches, `Kind.path`, and global `path` syntax - shown
    patching values in the already-rendered YAML without re-templating it.
  - Two `machine_pool` blocks (`control_plane`/`control_plane_as_worker` for the control-plane
    pool), each with `node_pool_config` rendered from
    [`config_templates/cp_pool_config.yaml`](config_templates/cp_pool_config.yaml) and
    [`config_templates/worker_pool_config.yaml`](config_templates/worker_pool_config.yaml), plus
    taints.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, and `namespaces`.
  - A settable `location_config` - unlike most other cluster resources, this is not
    Computed-only here, since Palette can't otherwise infer a custom cloud's physical location
    from raw CAPI YAML.
  - Top-level Day-2 settings: `cluster_type`, `pause_agent_upgrades`, OS patch scheduling,
    `cluster_timezone`, `update_worker_pools_in_parallel`, tags, description.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, and local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. A Private Cloud Gateway (PCG) already installed and reachable from your Nutanix Prism
   Central.
2. Nutanix Prism Central credentials, a target Prism Element cluster, a VM image template, and a
   subnet.
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
