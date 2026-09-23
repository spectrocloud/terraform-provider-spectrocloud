# spectrocloud_cluster_brownfield - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_brownfield` and its supporting resources in one working, self-contained
configuration. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_brownfield`](../../resources/spectrocloud_cluster_brownfield).

`spectrocloud_cluster_brownfield` is fundamentally different from every other cluster resource in
this repo: it doesn't provision any infrastructure. It **registers an already-running Kubernetes
cluster** with Palette for ongoing management (profiles, backup, RBAC, etc.). There is no cloud
account, no `cloud_config`, and no `kubeconfig`/`admin_kube_config` - instead, applying this
resource produces a `manifest_url` and `kubectl_command` that you must run against the real,
existing cluster to complete the import. Until that command is run, Palette shows the cluster as
pending import.

## What this demonstrates

- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using a self-hosted, S3-compatible object store (Minio), since a brownfield cluster
  can live on any infrastructure.
- **Cluster profile** (`clusterprofile.tf`): Kubernetes + CNI + one add-on pack (`cloud =
  "generic"`), and a `profile_variables` block covering all 4 variable formats. Brownfield
  profiles skip the OS and CSI layers other cloud types need, since Palette doesn't provision the
  underlying host or storage for a cluster that already exists.
- **Cluster** (`cluster.tf`):
  - `cloud_type`, `import_mode`, `context`, `proxy`/`no_proxy`, `host_path`/`container_mount_path`
    - all effectively set-once per the provider's own documentation, even though none are
    schema-level `ForceNew`.
  - Two `machine_pool` blocks used purely for Day-2 node maintenance (`node { action =
    "cordon"/"uncordon" }`) - unlike other cluster resources, `machine_pool` here never creates
    or sizes node pools.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, `cluster_timezone`, tags, description.
- **Local import-command export** (`import_command.tf`): writes the generated `kubectl_command`
  to disk on every apply, in place of the kubeconfig export the other examples do (brownfield
  clusters expose no kubeconfig of their own).
- **Outputs** (`outputs.tf`): cluster ID, `manifest_url`, `kubectl_command`, `status`,
  `health_status`, `cloud_config_id`, and the local import-command file path.

Day-2 mutability is documented inline throughout every `.tf` file.

## Prerequisites

1. An already-running Kubernetes cluster you want to bring under Palette management.
2. A Minio (or other S3-compatible) endpoint to use as the backup storage location.

## Usage

1. Copy `terraform.template.tfvars` to `terraform.tfvars` and fill in every placeholder value.
2. `terraform init && terraform apply`. This registers the cluster's *intent* with Palette; it
   does not yet complete the import.
3. Run the generated command against the real, existing cluster to finish the import:
   ```shell
   bash $(terraform output -raw import_command_path)
   ```
4. Check `terraform output status` / `terraform output health_status` to confirm Palette sees the
   cluster as imported and healthy.

## Clean up

```shell
terraform destroy
```

This stops Palette from managing the cluster; it does not affect the underlying cluster itself.
