# spectrocloud_cluster_aks - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_aks` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_aks`](../../resources/spectrocloud_cluster_aks).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_azure` resource (the same
  cloud account type used by `spectrocloud_cluster_azure`).
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using Azure Blob storage.
- **Cluster profile** (`clusterprofile.tf`): AKS-specific packs - `kubernetes-aks` (the
  control plane is Azure-managed, so this differs from the generic `kubernetes` pack used
  elsewhere) and `cni-kubenet` (Azure's native CNI) alongside `ubuntu-aks` and `csi-azure`, plus
  a `profile_variables` block covering all 4 variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field, including `private_cluster` and the full bring-your-own VNet
    path (a flat set of fields, not nested subnet blocks - AKS static placement doesn't support
    multiple subnets per role on the backend).
  - Three `machine_pool` blocks: a system node pool, a Linux worker pool with autoscaling
    (`min`/`max` + `override_scaling`), taints, labels/annotations, and
    `override_kubeadm_configuration`/`override_cluster_api_config`, and a Windows worker pool
    (`os_type = "Windows"`, `os_sku = "Windows2022"`). Note: unlike most other cluster
    resources here, machine_pool has no `control_plane`/`control_plane_as_worker` fields, since
    AKS's control plane is fully Azure-managed.
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

1. An Azure AD application (service principal) with permissions to create AKS clusters. See
   [Azure Cloud Account](https://docs.spectrocloud.com/clusters?clusterType=azure_cluster#creatinganazurecloudaccount).
2. An Azure storage account and blob container to use as the backup storage location.

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
