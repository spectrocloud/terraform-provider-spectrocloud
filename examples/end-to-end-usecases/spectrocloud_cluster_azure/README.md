# spectrocloud_cluster_azure - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_azure` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_azure`](../../resources/spectrocloud_cluster_azure).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_azure` resource with Azure
  AD application credentials, `cloud` partition selection, and the `tenant_name`/
  `disable_properties_request` fields.
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using the `azure_storage_config` block (Azure Blob storage), matching this cluster's
  cloud provider.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI (Azure-specific
  `cni-calico-azure`) + CSI + add-on pack stack, and a `profile_variables` block covering all 4
  variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field, including the full bring-your-own custom VNet path
    (`network_resource_group`/`virtual_network_name`/`virtual_network_cidr_block` plus
    `control_plane_subnet`/`worker_node_subnet`, which are `RequiredWith` each other) and the
    `private_api_server` custom-DNS option shown commented.
  - Three `machine_pool` blocks: a system-node-pool control plane, a Linux worker pool
    (taints, labels/annotations, `override_scaling`, `override_kubeadm_configuration`,
    `override_cluster_api_config`, `override_health_check_configuration`), and a Windows worker
    pool demonstrating `os_type = "Windows"`. Note: unlike some other cluster resources, this
    machine_pool has no `min`/`max` autoscaling fields.
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

1. An Azure AD application (service principal) with permissions to create the resources Palette
   provisions. See
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
