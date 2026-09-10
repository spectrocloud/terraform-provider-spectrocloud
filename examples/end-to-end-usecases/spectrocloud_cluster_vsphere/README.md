# spectrocloud_cluster_vsphere - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_vsphere` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones - so you can see every option in context and copy the pieces you need. For the
minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_vsphere`](../../resources/spectrocloud_cluster_vsphere).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_vsphere` resource is
  created here (every attribute: vCenter address, username/password, insecure-cert handling)
  instead of looking up a pre-existing account, so the whole example is self-contained.
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource is likewise created (not looked up), using AWS S3 with static access/secret key
  credentials; the GCP and Azure storage backends are shown as commented alternatives.
- **Cluster profile** (`clusterprofile.tf`): a full OS + Kubernetes + CNI + CSI + add-on pack
  stack, `context`, and a `profile_variables` block covering all 4 variable formats (hidden
  string, version with a regex constraint, dropdown, multiline). The advanced
  `skip_destroy`/immutable-versioning pattern is documented in a comment even though not used.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field populated (datacenter/folder/image template folder, SSH keys,
    DDNS networking with a commented static-IP alternative, NTP servers, CAPV YAML passthrough).
  - Two `machine_pool` blocks: a 3-node control plane, and a worker pool demonstrating
    autoscaling (`min`/`max` + `override_scaling`), taints, additional labels/annotations,
    `node_repave_interval`, `skip_k8s_upgrade`, and both `override_kubeadm_configuration` and
    `override_cluster_api_config` YAML passthroughs.
  - `cluster_profile` with per-cluster `variables` (feeding the profile's `profile_variables`)
    and a per-cluster `pack` override.
  - `backup_policy` (pointing at the backup storage location created above), `scan_policy`, two
    `cluster_rbac_binding` blocks (cluster-scoped and namespace-scoped), `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `pause_agent_upgrades`, OS patch scheduling, `cluster_timezone`,
    `update_worker_pools_in_parallel`, tags, description, and `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): uses the `hashicorp/local` provider to write
  both the regular and cluster-admin kubeconfigs to disk on every apply, instead of requiring a
  manual `terraform output -raw ... >` step.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfig/admin kubeconfig (also written to local
  files), the deprecated-but-still-populated `cloud_config_id`, the Computed-only
  `location_config`, and the local kubeconfig file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file - look for the "Day-2 mutability" comment at the top of each resource block.

## Prerequisites

1. A Private Cloud Gateway installed and reachable from your vCenter. See
   [VMware First Cluster](https://docs.spectrocloud.com/getting-started/?getting_started=vmware#yourfirstvmwarecluster).
   Its UID is `private_cloud_gateway_id` below - you do **not** need a pre-existing cloud account,
   this example creates one.
2. An AWS S3 bucket (or Minio/GCP/Azure equivalent, see the commented alternatives in
   `backup_storage.tf`) to use as the backup storage location.

## Usage

1. Copy `terraform.template.tfvars` to `terraform.tfvars` and fill in every placeholder value.
2. `terraform init && terraform apply`.
3. `kubeconfig.tf` writes the kubeconfig files to the current directory automatically. Point
   `kubectl` at the cluster:
   ```shell
   export KUBECONFIG=$(terraform output -raw kubeconfig_path)
   kubectl get pod -A
   ```

## Clean up

```shell
terraform destroy
```
