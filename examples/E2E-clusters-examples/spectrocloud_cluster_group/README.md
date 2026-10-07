# spectrocloud_cluster_group - comprehensive end-to-end usecase

This is a **kitchen-sink** example: it exercises essentially the full attribute surface of
`spectrocloud_cluster_group` and its supporting resources in one working, self-contained
configuration. For the minimal happy-path version of just the group resource, see
[`examples/Cluster Group/resource`](../../Cluster%20Group/resource).

`spectrocloud_cluster_group` is fundamentally different from the cloud cluster resources in this
repo: it has no cloud account and provisions no infrastructure of its own. It's a "hostCluster"
type overlay that logically groups one or more already-existing Palette-managed clusters, so that
[virtual clusters](../spectrocloud_virtual_cluster) (via their `cluster_group_uid` attribute) can
be scheduled across the group collectively instead of being pinned to one specific
`host_cluster_uid`.

## What this demonstrates

- **Cluster profile** (`clusterprofile.tf`): an `add-on`-type profile (`cloud = "all"`) carrying a
  manifest pack and a `profile_variables` block covering all 4 variable formats - attached to the
  group as a whole so it applies across every host cluster in it.
- **Cluster group** (`cluster_group.tf`):
  - `tags`, `description`, and `context` (tenant-scoped, the default for cluster groups).
  - `config`: every resource-quota field (`cpu_millicore`, `memory_in_mb`, `storage_in_gb`,
    `oversubscription_percent`), `host_endpoint_type` set to the non-default `"LoadBalancer"`, a
    YAML `values` override, and `k8s_distribution` set to `"cncf_k8s"` (the deprecated `"k3s"`
    alternative is called out in a comment, not used).
  - `cluster_profile` with group-level `variables` and a group-level `pack` override - same block
    shape used by `spectrocloud_virtual_cluster`.
  - Two `clusters` blocks (`cluster_uid` + `host_dns` each) to exercise the repeatable block - a
    single host cluster is just as valid.
- **Outputs** (`outputs.tf`): the group ID and its add-on profile ID. Unlike the cloud cluster
  resources in this repo, a cluster group has no kubeconfig of its own, so there's no
  `kubeconfig.tf` here.

Day-2 mutability (ForceNew vs. updatable in place) is documented inline throughout every `.tf`
file.

## Prerequisites

1. Two existing Palette-managed clusters to add as host clusters (their UIDs are
   `host_cluster_uid_primary` / `host_cluster_uid_secondary`). A single host cluster also works -
   just remove the second `clusters` block in `cluster_group.tf` and the corresponding variables.

## Usage

1. Copy `terraform.template.tfvars` to `terraform.tfvars` and fill in every placeholder value.
2. `terraform init && terraform apply`.
3. Optionally, deploy a `spectrocloud_virtual_cluster` with
   `cluster_group_uid = spectrocloud_cluster_group.cg.id` to see Palette schedule it onto one of
   the group's host clusters - see
   [`examples/E2E-clusters-examples/spectrocloud_virtual_cluster`](../spectrocloud_virtual_cluster).

## Clean up

```shell
terraform destroy
```
