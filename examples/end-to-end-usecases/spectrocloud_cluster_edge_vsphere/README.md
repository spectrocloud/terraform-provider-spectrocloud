# spectrocloud_cluster_edge_vsphere - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_edge_vsphere` and its supporting resources in one working, self-contained
configuration. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_edge_vsphere`](../../resources/spectrocloud_cluster_edge_vsphere).

## What this demonstrates

- **Edge appliance** (`edgehost.tf`): a `spectrocloud_appliance` resource, registering a
  pre-paired physical/virtual edge host (running a vSphere-compatible virtualization layer) with
  Palette, rather than looking up one paired out-of-band. Unlike the plain (public-cloud)
  `spectrocloud_cluster_vsphere`, there is no cloud account here - the whole cluster runs on this
  single appliance, referenced by the cluster's top-level `edge_host_uid`.
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio) - the common choice for
  on-prem/edge clusters, alongside the AWS/GCP/Azure alternatives shown commented.
- **Cluster profile** (`clusterprofile.tf`): Kubernetes + CNI + one add-on pack, and a
  `profile_variables` block covering all 4 variable formats. Like Edge Native, Edge vSphere
  profiles skip the OS and CSI layers other cloud types need.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field: `datacenter`, `folder`, `image_template_folder`, `ssh_keys`,
    `vip`, `static_ip`, `network_type`, `network_search_domain`.
  - Two `machine_pool` blocks with vSphere `placement` (cluster/resource_pool/datastore/network)
    and `instance_type` (disk_size_gb/memory_mb/cpu), taints, labels/annotations,
    `override_scaling`, and the kubeadm/health-check YAML passthroughs. Unlike plain vSphere,
    there is no `min`/`max` autoscaling here - pool size is a fixed `count`.
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

1. An edge appliance (running a vSphere-compatible virtualization layer) already paired with
   Palette - see [Edge vSphere Deployment](https://docs.spectrocloud.com/clusters/edge/).
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
