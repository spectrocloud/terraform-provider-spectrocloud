# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

Most resource/data-source examples are organized directly under this directory to mirror Palette's own left-nav product structure, instead of the flat `<full resource name>` folders `tfplugindocs` uses by default. It's meant to be browsed: if you know where something lives in the Palette UI, it lives in the same place here (`Projects/`, `Clusters/`, `Infrastructure/`, `Security/`, `Platform/`, etc. - see "Navigation structure" below for the full tree).

Each leaf folder is a standalone, runnable example: `providers.tf`, `resource.tf` or `data-source.tf`, and usually `terraform.template.tfvars` (copy to `terraform.tfvars` and fill in real values to run it - the real file is gitignored).

### Contents

* **provider/provider.tf** example file for the provider index page
* **Projects/, Cluster Profiles/, Clusters/, Infrastructure/, Security/, Platform/, App Mode/, Roles/, Users/, Teams/, Cluster Group/, Cluster Configuration/, Virtual Machines/, Workspaces/** - the navigation-aligned examples; see "Navigation structure" below.
* **Other/** - resources/data sources that don't map cleanly onto a single Palette nav entry; see below.
* **E2E-clusters-examples/\<resource\>/** one comprehensive, "kitchen-sink" example per cluster-provisioning resource, demonstrating end-to-end cluster provisioning - see [E2E-clusters-examples/README.md](E2E-clusters-examples/README.md).
* **imports/** a single reference file (`import.tf`) documenting `terraform import` for every importable resource - see [imports/README.md](imports/README.md).
* **Spectro-TF-tutorials/\<name\>/** directory contains examples aligned to specific tutorials at [docs.spectrocloud.com/tutorials](https://docs.spectrocloud.com/tutorials/) — see [Spectro-TF-tutorials/README.md](Spectro-TF-tutorials/README.md) for the full list.

### How docs generation finds examples

`tfplugindocs` (run via `make generate`) only auto-discovers examples at the fixed path
`examples/{resources,data-sources}/<full name>/{resource,data-source}.tf`, which this directory no
longer uses. Every resource/data source instead has a matching template at
`templates/{resources,data-sources}/<name>.md.tmpl` (name without the `spectrocloud_` prefix) whose
`{{ tffile "..." }}` call points at its file here. If you move a folder in this tree, update the
`tffile` path in the same template - otherwise `make generate` keeps rendering the old content, or
silently falls back to no example at all.

## Navigation structure

```
Projects/                              spectrocloud_project
  Project Settings/
    Alert/                              spectrocloud_alert
Cluster Profiles/                      spectrocloud_cluster_profile
  resource/Helm/                       spectrocloud_cluster_profile (OCI Helm chart example)
  resource/Import/                     spectrocloud_cluster_profile_import
Cluster Configuration/
  Cluster Policy/                      spectrocloud_cluster_config_policy
  Cluster Template/                    spectrocloud_cluster_config_template
Clusters/
  data-source/                         spectrocloud_cluster
  data-source/Appliance/                spectrocloud_appliance (lookup)
  data-source/Appliances/               spectrocloud_appliances (filtered list of IDs)
  resource/
    AWS IaaS, AWS EKS, Azure IaaS, Azure AKS, GCP IaaS, GCP GKE,
    VMware vSphere, MAAS, Apache CloudStack, Edge Native, Custom Cloud,
    Edge VMware                        spectrocloud_cluster_edge_vsphere
    Brownfield                         spectrocloud_cluster_brownfield
    Appliance                          spectrocloud_appliance
    Addon Deployment                   spectrocloud_addon_deployment
Cluster Group/                         spectrocloud_cluster_group
Roles/
  data-source/                         spectrocloud_role (lookup)
  data-source/Permission/               spectrocloud_permission (lookup)
  resource/
    Project Roles, Tenant Roles, Resource Roles     spectrocloud_role (per scope)
Users/                                 spectrocloud_user
Teams/                                 spectrocloud_team
Infrastructure/
  Cloud Accounts/
    data-source/ and resource/, split per cloud: AWS, Azure, Google Cloud,
    VMware vSphere, MAAS, Apache CloudStack, Custom Cloud
  Cloud Rates/                         coming soon (no backing resource/data source yet)
  Registries/
    data-source/                       spectrocloud_registry (generic, any type)
    data-source/ and resource/, split: Helm Registries, OCI Registries
    data-source/Pack Registries         spectrocloud_registry_pack (lookup only)
    resource/Pack Registries            coming soon (no backing resource yet)
    data-source/Pack                    spectrocloud_pack (lookup within a registry)
    data-source/Pack Simple             spectrocloud_pack_simple (simpler lookup)
  Private Cloud Gateway/
    data-source/                       spectrocloud_private_cloud_gateway
    data-source/DNS Map/                spectrocloud_privatecloudgateway_dns_map (lookup)
    data-source/IP Pool/                 spectrocloud_ippool (lookup)
    resource/DNS Map/                   spectrocloud_privatecloudgateway_dns_map
    resource/IP Pool/                   spectrocloud_privatecloudgateway_ippool
  Backup Location/                     spectrocloud_backup_storage_location
  Audit Trails/                        spectrocloud_audit_trail (resource only)
Security/
  SSH Keys/                            spectrocloud_ssh_key
  SSO/
    SAML/, OIDC/                       spectrocloud_sso (per protocol)
  Password Policy/                     spectrocloud_password_policy
  Registration Token/                  spectrocloud_registration_token (resource + data-source)
  Hardened Images/                     coming soon (no backing resource/data source yet)
  Developer Settings/                  spectrocloud_developer_setting
Platform/
  Certificates/                        coming soon (no backing resource/data source yet)
  Macros/                              spectrocloud_macros (resource + data-source)
  Resource Limits/                     spectrocloud_resource_limit
  Platform Settings/                   spectrocloud_platform_setting
  Filters/                             spectrocloud_filter (resource + data-source)
App Mode/
  App Profiles/                        spectrocloud_application_profile (data-source + resource)
  Apps/                                spectrocloud_application (resource only)
  Virtual Clusters/                    spectrocloud_virtual_cluster (resource only)
Virtual Machines/
  resource/                            spectrocloud_virtual_machine
  resource/Data Volume/                spectrocloud_datavolume
Workspaces/                            spectrocloud_workspace
Other/                                 see below
```

### "coming soon" stubs

`Infrastructure/Cloud Rates/`, `Infrastructure/Registries/resource/Pack Registries/`, and
`Security/Hardened Images/`, `Platform/Certificates/` appear in Palette's nav but have no
corresponding provider resource/data source today. Each is an empty folder with a `README.md`
explaining why, so the navigation tree stays complete and future resources have an obvious home.

### `Other/`

Only `Other/cloud_account/` remains, and it's a flag, not a real placeholder: there is no
`spectrocloud_cloud_account` registered in `provider.go` - only the per-cloud
`spectrocloud_cloudaccount_<aws|azure|gcp|...>` types, already under `Infrastructure/Cloud
Accounts/`. Its `data-source.tf` was byte-identical to the AWS cloud account lookup already at
`Infrastructure/Cloud Accounts/data-source/AWS/data-source.tf`, so it looks like orphaned, dead
content left over from before the per-cloud split - it was intentionally left in place rather
than moved or deleted, pending a decision on whether to remove it outright.
