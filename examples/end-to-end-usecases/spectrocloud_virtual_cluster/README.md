# spectrocloud_virtual_cluster - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_virtual_cluster` and its supporting resources in one working, self-contained
configuration. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_virtual_cluster`](../../resources/spectrocloud_virtual_cluster).

`spectrocloud_virtual_cluster` is fundamentally different from the other cloud cluster resources
in this repo: it has no cloud account and provisions no infrastructure of its own. It's a
Kubernetes-in-Kubernetes control plane (vcluster) that runs inside an already-existing
Palette-managed host cluster (or whichever cluster a cluster group selects).

## What this demonstrates

- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio), since a virtual cluster's
  host cluster can be on any infrastructure.
- **Cluster profile** (`clusterprofile.tf`): an `add-on`-type profile (`cloud = "all"`) carrying
  a manifest pack and a `profile_variables` block covering all 4 variable formats. Virtual
  clusters get their Kubernetes control plane from the vcluster Helm chart (`cloud_config` on the
  cluster resource), not from an OS/kubernetes/cni/csi pack stack.
- **Cluster** (`cluster.tf`):
  - `host_cluster_uid` (this example) vs. the commented `cluster_group_uid` alternative -
    exactly one places the virtual cluster.
  - `pause_cluster` and a `resources` block bounding CPU/memory/storage (all 6 fields optional).
  - `cloud_config`: the Helm chart (`chart_name`/`chart_repo`/`chart_version`/`chart_values`) and
    `k8s_version` used to install the vcluster control plane - the only part of this resource
    that's ForceNew.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, and `namespaces`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, `apply_setting`,
    `update_worker_pools_in_parallel`, `cluster_timezone`, OS patch scheduling, tags, description.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, and local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. An existing Palette-managed cluster to host the virtual cluster (its UID is
   `host_cluster_uid`).
2. A Minio (or other S3-compatible) endpoint to use as the backup storage location.

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
