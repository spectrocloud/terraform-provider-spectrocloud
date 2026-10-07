# imports - import examples for every supported resource

`import.tf` is a single reference file with one `terraform import`-style example per
`spectrocloud_*` resource that supports import (i.e. every resource with an `Importer` defined
in its Go schema) - 51 resources in total, grouped into sections:

- Cluster resources (`spectrocloud_cluster_aws`, `..._azure`, `..._gcp`, `..._vsphere`,
  `..._aks`, `..._eks`, `..._gke`, `..._maas`, `..._edge_native`, `..._edge_vsphere`,
  `..._apache_cloudstack`, `..._custom_cloud`, `..._brownfield`, `spectrocloud_virtual_cluster`)
- Cluster-adjacent resources (`cluster_group`, `cluster_profile`, `cluster_config_template`,
  `cluster_config_policy`)
- Cloud accounts (`cloudaccount_aws`, `..._azure`, `..._gcp`, `..._vsphere`, `..._maas`,
  `..._apache_cloudstack`, `..._custom`)
- Backup storage & private cloud gateway infrastructure (`backup_storage_location`,
  `appliance`, `privatecloudgateway_ippool`, `privatecloudgateway_dns_map`)
- IAM & organization (`role`, `team`, `user`, `ssh_key`, `project`, `workspace`, `filter`,
  `registration_token`)
- Pack registries (`registry_helm`, `registry_oci`)
- Tenant-level settings (`resource_limit`, `password_policy`, `developer_setting`,
  `platform_setting`, `macros`, `sso`)
- Alerts & audit (`alert`, `audit_trail`)
- Applications (`application`, `application_profile`)
- KubeVirt (`virtual_machine`, `datavolume`)

## What each entry shows

For every resource:

1. A comment documenting its exact import ID format - most Palette resources use
   `<uid_or_name>:<project|tenant>` (the first segment is tried as a UID, then falls back to a
   name lookup within that context), but several resources deviate from this (an extra segment,
   no context suffix at all, or a different separator entirely) - those are called out
   explicitly.
2. The equivalent classic CLI command, as a comment: `terraform import <address> <id>`.
3. The config-generating command pair, as comments:
   `terraform plan -generate-config-out=generated_<resource>.tf` followed by `terraform apply`.
4. A config-driven `import` block (Terraform >= 1.5):
   ```hcl
   import {
     to = spectrocloud_cluster_aws.example
     id = "my-aws-cluster:project"
   }
   ```

This file is reference material, not a working configuration - `to` doesn't need a matching
`resource` block for `terraform validate` to pass, which is why the whole file validates cleanly
as-is. To actually run one of these imports, pick one of three equivalent workflows:

- **Generate the config for you (recommended if you don't have a `resource` block yet).** Copy
  the `import` block for your resource into your configuration - no `resource` block needed -
  then run:
  ```shell
  terraform plan -generate-config-out=generated_<resource>.tf
  ```
  This fetches the real object from Palette and writes a starting `resource "<type>" "<name>"
  { ... }` block into `generated_<resource>.tf` for you. Review/edit it (Terraform can't always
  infer every attribute perfectly, e.g. sensitive fields), then run `terraform apply` to write
  the import to state.
- **Reconcile against a config you already wrote.** If you already have a `resource` block
  matching `to`, copy the `import` block alongside it and run plain `terraform plan` /
  `terraform apply` (no `-generate-config-out`) - Terraform reconciles the import against your
  existing config instead of generating a new file.
- **Classic CLI, Terraform < 1.5.** Add a `resource "<type>" "<name>" { ... }` block yourself
  first (see `examples/resources/<type>` for a minimal starting point), then run the commented
  `terraform import <address> <id>` command - this only writes state, it never generates config.

After any of these, run `terraform plan` again - if your resource block doesn't yet match every
attribute Palette reports, Terraform shows a diff you'll need to reconcile before applying
normally.

## Usage

```shell
terraform init
terraform validate
```

There's nothing to `apply` here - this folder exists to document import ID formats, not to
provision anything. `sc_api_key` defaults to an empty string in `providers.tf`, so no
`terraform.tfvars` is required just to validate; set real credentials only if you copy an
`import` block into a configuration you intend to actually run.
