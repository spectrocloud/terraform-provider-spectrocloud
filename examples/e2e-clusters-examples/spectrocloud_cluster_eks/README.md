# spectrocloud_cluster_eks - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_eks` and its supporting resources in one working, self-contained
configuration - it creates its own cloud account and backup storage location rather than looking
up existing ones. For the minimal happy-path version of just the cluster resource, see
[`examples/resources/spectrocloud_cluster_eks`](../../resources/spectrocloud_cluster_eks).

## What this demonstrates

- **Cloud account** (`cloudaccount.tf`): a `spectrocloud_cloudaccount_aws` resource; the
  EKS-relevant "pod-identity" credential type is shown commented alongside "sts".
- **Backup storage location** (`backup_storage.tf`): a `spectrocloud_backup_storage_location`
  resource using AWS S3.
- **Cluster profile** (`clusterprofile.tf`): EKS-specific packs - `amazon-linux-eks`,
  `kubernetes-eks`, and `cni-aws-vpc-eks` (all distinct from the packs used by the generic
  `spectrocloud_cluster_aws` resource) plus `csi-aws`, and a `profile_variables` block covering
  all 4 variable formats.
- **Cluster** (`cluster.tf`):
  - Every `cloud_config` field, including `endpoint_access` and the public/private access CIDR
    lists - the only two fields in `cloud_config` that update in place rather than ForceNew.
  - One `machine_pool` (EKS Managed Node Group) demonstrating autoscaling (note: EKS requires
    `count` to equal `min` when autoscaling is enabled), spot capacity, taints,
    labels/annotations, `ami_type`, `override_scaling`,
    `override_kubeadm_configuration`/`override_cluster_api_config`, and a full
    `eks_launch_template` block. Note: like AKS, machine_pool has no
    `control_plane`/`control_plane_as_worker` fields, since EKS's control plane is fully
    AWS-managed.
  - A `fargate_profile` block - EKS-specific serverless compute, running pods matching a
    namespace/label selector without a backing machine pool.
  - `cluster_profile` with per-cluster `variables` and a per-cluster `pack` override.
  - `backup_policy`, `scan_policy`, two `cluster_rbac_binding` blocks, `namespaces`, and
    `host_config`.
  - Top-level Day-2 settings: `tags_map`, `pause_agent_upgrades`, OS patch scheduling,
    `cluster_timezone`, `update_worker_pools_in_parallel`, description, `cluster_meta_attribute`.
- **Local kubeconfig export** (`kubeconfig.tf`): writes both kubeconfigs to disk on every apply.
- **Outputs** (`outputs.tf`): cluster ID, kubeconfigs, `cloud_config_id`, and local file paths.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. An AWS account with permissions to create EKS clusters and managed node groups. See
   [AWS Cloud Account](https://docs.spectrocloud.com/clusters/?clusterType=aws_cluster#awscloudaccountpermissions).
2. An existing EC2 key pair, and subnet IDs if you want to exercise the Fargate profile.
3. An S3 bucket to use as the backup storage location.

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
