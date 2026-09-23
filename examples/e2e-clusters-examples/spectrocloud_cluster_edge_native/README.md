# spectrocloud_cluster_edge_native - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_edge_native` and its supporting resources in one working, self-contained
configuration. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_edge_native`](../../resources/spectrocloud_cluster_edge_native).

## What this demonstrates

- **Edge appliances** (`edgehost.tf`): two `spectrocloud_appliance` resources (control-plane and
  worker), registering pre-paired physical/virtual edge hosts with Palette rather than looking up
  ones paired out-of-band. Edge Native clusters have no cloud account concept - nodes are
  provisioned via these appliance pairings instead.
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio) - the common choice for
  on-prem/edge clusters, alongside the AWS/GCP/Azure alternatives shown commented.
- **Cluster profile** (`clusterprofile.tf`): Kubernetes + CNI + one add-on pack, and a
  `profile_variables` block covering all 4 variable formats. Edge Native profiles skip the OS and
  CSI layers other cloud types need - edge appliances bring their own OS image (BYOI), and
  on-prem storage is typically handled outside a Palette-managed CSI pack.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field: `ssh_keys`, `vip`, `overlay_cidr_range`, `is_two_node_cluster`,
    `ntp_servers`.
  - Two `machine_pool` blocks, each with a required `edge_host` block mapping the pool onto a
    registered appliance, plus its full networking override surface (`static_ip`,
    `default_gateway`, `subnet_mask`, `dns_servers`, `nic_name`, `two_node_role`), taints,
    labels/annotations, `override_scaling`, and the kubeadm/health-check YAML passthroughs.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    `update_worker_pools_in_parallel`, tags, description, `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, `location_config`, and
  local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. Two edge appliances (physical or virtual) already paired with Palette - see
   [Edge Native Deployment](https://docs.spectrocloud.com/clusters/edge/).
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
