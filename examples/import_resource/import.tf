# Import examples for every spectrocloud_* resource that supports `terraform import`
# (i.e. every resource with an Importer defined in its Go schema).
#
# This single file is reference material, not a working configuration: each block below shows
# how to bring one existing Palette object under Terraform management, in two equivalent forms:
#
#   1. A config-driven `import` block (Terraform >= 1.5). `to` is the resource address you will
#      declare in your own configuration; `id` is this resource's import ID, in the exact format
#      documented in the comment above each block. Run `terraform plan` to preview the import and
#      `terraform apply` to write it to state - both require a matching `resource "<type>" "<name>"`
#      block to already exist in your configuration (see examples/resources/<type> for one).
#   2. The equivalent classic `terraform import <address> <id>` CLI command, shown commented
#      directly below each block - use this if you're on Terraform < 1.5, or prefer the one-shot
#      CLI workflow over a plan/apply cycle.
#
# General ID convention: most Palette resources are scoped to a `project` or `tenant` context,
# and their import ID is `<uid_or_name>:<context>` - the first segment is tried as a UID first,
# then falls back to a name lookup within that context. Resources that deviate from this (no
# context suffix, an extra segment, a different separator) are called out explicitly in their
# section below.
#
# `terraform validate` does not require the `to` address to have a matching `resource` block, so
# this file validates cleanly on its own - but a real `plan`/`apply` against these `import` blocks
# does need one, added to your own configuration.


##############################################################################
# Cluster resources
##############################################################################
# All spectrocloud_cluster_* resources below (except where noted) share the same import ID
# format: `<cluster_id_or_name>:<project|tenant>`. The first segment is tried as the cluster UID
# first, then as the cluster name within the given context.

# --- spectrocloud_cluster_aws ---
# terraform import spectrocloud_cluster_aws.example my-aws-cluster:project
import {
  to = spectrocloud_cluster_aws.example
  id = "my-aws-cluster:project"
}

# --- spectrocloud_cluster_azure ---
# terraform import spectrocloud_cluster_azure.example my-azure-cluster:project
import {
  to = spectrocloud_cluster_azure.example
  id = "my-azure-cluster:project"
}

# --- spectrocloud_cluster_gcp ---
# terraform import spectrocloud_cluster_gcp.example my-gcp-cluster:project
import {
  to = spectrocloud_cluster_gcp.example
  id = "my-gcp-cluster:project"
}

# --- spectrocloud_cluster_vsphere ---
# terraform import spectrocloud_cluster_vsphere.example my-vsphere-cluster:project
import {
  to = spectrocloud_cluster_vsphere.example
  id = "my-vsphere-cluster:project"
}

# --- spectrocloud_cluster_aks ---
# terraform import spectrocloud_cluster_aks.example my-aks-cluster:project
import {
  to = spectrocloud_cluster_aks.example
  id = "my-aks-cluster:project"
}

# --- spectrocloud_cluster_eks ---
# terraform import spectrocloud_cluster_eks.example my-eks-cluster:project
import {
  to = spectrocloud_cluster_eks.example
  id = "my-eks-cluster:project"
}

# --- spectrocloud_cluster_gke ---
# terraform import spectrocloud_cluster_gke.example my-gke-cluster:project
import {
  to = spectrocloud_cluster_gke.example
  id = "my-gke-cluster:project"
}

# --- spectrocloud_cluster_maas ---
# terraform import spectrocloud_cluster_maas.example my-maas-cluster:project
import {
  to = spectrocloud_cluster_maas.example
  id = "my-maas-cluster:project"
}

# --- spectrocloud_cluster_edge_native ---
# terraform import spectrocloud_cluster_edge_native.example my-edge-native-cluster:project
import {
  to = spectrocloud_cluster_edge_native.example
  id = "my-edge-native-cluster:project"
}

# --- spectrocloud_cluster_edge_vsphere ---
# terraform import spectrocloud_cluster_edge_vsphere.example my-edge-vsphere-cluster:project
import {
  to = spectrocloud_cluster_edge_vsphere.example
  id = "my-edge-vsphere-cluster:project"
}

# --- spectrocloud_cluster_apache_cloudstack ---
# terraform import spectrocloud_cluster_apache_cloudstack.example my-cloudstack-cluster:project
import {
  to = spectrocloud_cluster_apache_cloudstack.example
  id = "my-cloudstack-cluster:project"
}

# --- spectrocloud_cluster_custom_cloud ---
# Deviates from the common format: takes a THIRD segment naming the custom cloud provider
# (e.g. "nutanix"), which is set on the resource's `cloud` attribute on import.
# Format: <cluster_id_or_name>:<project|tenant>:<custom_cloud_name>
# terraform import spectrocloud_cluster_custom_cloud.example my-nutanix-cluster:project:nutanix
import {
  to = spectrocloud_cluster_custom_cloud.example
  id = "my-nutanix-cluster:project:nutanix"
}

# --- spectrocloud_cluster_brownfield ---
# Deviates from the common format: takes a THIRD segment naming the cloud type the imported
# cluster actually runs on (aws, azure, gcp, generic, apache-cloudstack), which is set on the
# resource's `cloud_type` attribute on import.
# Format: <cluster_id_or_name>:<project|tenant>:<cloud_type>
# terraform import spectrocloud_cluster_brownfield.example my-existing-cluster:project:generic
import {
  to = spectrocloud_cluster_brownfield.example
  id = "my-existing-cluster:project:generic"
}

# --- spectrocloud_virtual_cluster ---
# Deviates from the common format: NO context suffix - virtual clusters always import into the
# "project" context. The ID is just the cluster's UID or name.
# Format: <virtual_cluster_id_or_name>
# terraform import spectrocloud_virtual_cluster.example my-virtual-cluster
import {
  to = spectrocloud_virtual_cluster.example
  id = "my-virtual-cluster"
}


##############################################################################
# Cluster-adjacent resources (profiles, templates, policies, groups)
##############################################################################

# --- spectrocloud_cluster_group ---
# Format: <cluster_group_id_or_name>:<project|tenant>
# terraform import spectrocloud_cluster_group.example my-cluster-group:tenant
import {
  to = spectrocloud_cluster_group.example
  id = "my-cluster-group:tenant"
}

# --- spectrocloud_cluster_profile ---
# Format: <profile_id_or_name>:<project|tenant|system>[:<version>]
# The version segment is optional; when omitted, the profile's current/latest version is used.
# terraform import spectrocloud_cluster_profile.example my-profile:project:1.0.0
import {
  to = spectrocloud_cluster_profile.example
  id = "my-profile:project:1.0.0"
}

# --- spectrocloud_cluster_config_template (Tech Preview) ---
# Format: <template_id_or_name>:<project|tenant>
# terraform import spectrocloud_cluster_config_template.example my-template:project
import {
  to = spectrocloud_cluster_config_template.example
  id = "my-template:project"
}

# --- spectrocloud_cluster_config_policy (Tech Preview) ---
# Format: <policy_id_or_name>:<project|tenant>
# terraform import spectrocloud_cluster_config_policy.example my-config-policy:project
import {
  to = spectrocloud_cluster_config_policy.example
  id = "my-config-policy:project"
}


##############################################################################
# Cloud accounts
##############################################################################
# All spectrocloud_cloudaccount_* resources below (except custom) share the same import ID
# format: `<account_id_or_name>:<project|tenant>`.

# --- spectrocloud_cloudaccount_aws ---
# terraform import spectrocloud_cloudaccount_aws.example my-aws-account:project
import {
  to = spectrocloud_cloudaccount_aws.example
  id = "my-aws-account:project"
}

# --- spectrocloud_cloudaccount_azure ---
# terraform import spectrocloud_cloudaccount_azure.example my-azure-account:project
import {
  to = spectrocloud_cloudaccount_azure.example
  id = "my-azure-account:project"
}

# --- spectrocloud_cloudaccount_gcp ---
# terraform import spectrocloud_cloudaccount_gcp.example my-gcp-account:project
import {
  to = spectrocloud_cloudaccount_gcp.example
  id = "my-gcp-account:project"
}

# --- spectrocloud_cloudaccount_vsphere ---
# terraform import spectrocloud_cloudaccount_vsphere.example my-vsphere-account:project
import {
  to = spectrocloud_cloudaccount_vsphere.example
  id = "my-vsphere-account:project"
}

# --- spectrocloud_cloudaccount_maas ---
# terraform import spectrocloud_cloudaccount_maas.example my-maas-account:project
import {
  to = spectrocloud_cloudaccount_maas.example
  id = "my-maas-account:project"
}

# --- spectrocloud_cloudaccount_apache_cloudstack ---
# terraform import spectrocloud_cloudaccount_apache_cloudstack.example my-cloudstack-account:project
import {
  to = spectrocloud_cloudaccount_apache_cloudstack.example
  id = "my-cloudstack-account:project"
}

# --- spectrocloud_cloudaccount_custom ---
# Deviates from the common format: takes a THIRD segment naming the custom cloud provider
# (e.g. "nutanix"), which is set on the resource's `cloud` attribute on import.
# Format: <account_id_or_name>:<project|tenant>:<custom_cloud_name>
# terraform import spectrocloud_cloudaccount_custom.example my-nutanix-account:project:nutanix
import {
  to = spectrocloud_cloudaccount_custom.example
  id = "my-nutanix-account:project:nutanix"
}


##############################################################################
# Backup storage & private cloud gateway infrastructure
##############################################################################

# --- spectrocloud_backup_storage_location ---
# Format: <bsl_id_or_name>[:<project|tenant>] - the context suffix is optional and defaults to
# "project" when omitted.
# terraform import spectrocloud_backup_storage_location.example my-backup-location:project
import {
  to = spectrocloud_backup_storage_location.example
  id = "my-backup-location:project"
}

# --- spectrocloud_appliance ---
# Format: <appliance_uid_or_name> - always project-scoped, no context suffix.
# terraform import spectrocloud_appliance.example my-edge-appliance-uid
import {
  to = spectrocloud_appliance.example
  id = "my-edge-appliance-uid"
}

# --- spectrocloud_privatecloudgateway_ippool ---
# Deviates from the common format: a composite key of the PARENT PCG plus the IP pool itself,
# not a context suffix. Either segment can be a UID or a name, in any combination.
# Format: <pcg_id_or_name>:<ip_pool_id_or_name>
# terraform import spectrocloud_privatecloudgateway_ippool.example my-pcg:my-ip-pool
import {
  to = spectrocloud_privatecloudgateway_ippool.example
  id = "my-pcg:my-ip-pool"
}

# --- spectrocloud_privatecloudgateway_dns_map ---
# Deviates from the common format: a composite key of the PARENT PCG plus the DNS map itself,
# not a context suffix. Either segment can be a UID or a name, in any combination.
# Format: <pcg_id_or_name>:<dns_map_id_or_name>
# terraform import spectrocloud_privatecloudgateway_dns_map.example my-pcg:my-dns-map
import {
  to = spectrocloud_privatecloudgateway_dns_map.example
  id = "my-pcg:my-dns-map"
}


##############################################################################
# IAM & organization
##############################################################################

# --- spectrocloud_role ---
# Format: <role_uid_or_name> - always tenant-scoped, no context suffix.
# terraform import spectrocloud_role.example my-custom-role
import {
  to = spectrocloud_role.example
  id = "my-custom-role"
}

# --- spectrocloud_team ---
# Deviates from the common pattern: UID ONLY - unlike most other resources, there is no
# import-by-name fallback for teams.
# Format: <team_uid>
# terraform import spectrocloud_team.example 5f6e7d8c9b0a1234567890ab
import {
  to = spectrocloud_team.example
  id = "5f6e7d8c9b0a1234567890ab"
}

# --- spectrocloud_user ---
# Format: <user_uid_or_email> - always tenant-scoped, no context suffix.
# terraform import spectrocloud_user.example jane.doe@example.com
import {
  to = spectrocloud_user.example
  id = "jane.doe@example.com"
}

# --- spectrocloud_ssh_key ---
# Format: <ssh_key_id_or_name>[:<project|tenant>] - the context suffix is optional and defaults
# to "project" when omitted.
# terraform import spectrocloud_ssh_key.example my-ssh-key:project
import {
  to = spectrocloud_ssh_key.example
  id = "my-ssh-key:project"
}

# --- spectrocloud_project ---
# Format: <project_uid_or_name> - projects are themselves the context, so there is no context
# suffix.
# terraform import spectrocloud_project.example my-project-name
import {
  to = spectrocloud_project.example
  id = "my-project-name"
}

# --- spectrocloud_workspace ---
# Format: <workspace_uid_or_name> - scoped to the provider's configured project, no context
# suffix.
# terraform import spectrocloud_workspace.example my-workspace
import {
  to = spectrocloud_workspace.example
  id = "my-workspace"
}

# --- spectrocloud_filter ---
# Format: <filter_uid_or_name> - scoped to the provider's configured project, no context suffix.
# terraform import spectrocloud_filter.example my-tag-filter
import {
  to = spectrocloud_filter.example
  id = "my-tag-filter"
}

# --- spectrocloud_registration_token ---
# Format: <token_uid_or_name> - always tenant-scoped, no context suffix.
# terraform import spectrocloud_registration_token.example my-registration-token
import {
  to = spectrocloud_registration_token.example
  id = "my-registration-token"
}


##############################################################################
# Pack registries
##############################################################################

# --- spectrocloud_registry_helm ---
# Format: <registry_uid_or_name> - always tenant-scoped, no context suffix.
# terraform import spectrocloud_registry_helm.example my-helm-registry
import {
  to = spectrocloud_registry_helm.example
  id = "my-helm-registry"
}

# --- spectrocloud_registry_oci_ecr ---
# Format: <registry_uid_or_name> - always tenant-scoped, no context suffix. Resolves against
# both ECR-backed and generic OCI-basic-auth registries.
# terraform import spectrocloud_registry_oci_ecr.example my-ecr-registry
import {
  to = spectrocloud_registry_oci_ecr.example
  id = "my-ecr-registry"
}


##############################################################################
# Tenant-level settings
##############################################################################

# --- spectrocloud_resource_limit ---
# Singleton resource - one instance per tenant.
# Format: <tenant_uid_or_org_name> - must match the tenant your provider is authenticated
# against; the provider rejects an import for any other tenant.
# terraform import spectrocloud_resource_limit.example my-org-name
import {
  to = spectrocloud_resource_limit.example
  id = "my-org-name"
}

# --- spectrocloud_password_policy ---
# Singleton resource - one instance per tenant.
# Format: <tenant_uid_or_org_name> - must match the tenant your provider is authenticated
# against.
# terraform import spectrocloud_password_policy.example my-org-name
import {
  to = spectrocloud_password_policy.example
  id = "my-org-name"
}

# --- spectrocloud_developer_setting ---
# Singleton resource - one instance per tenant.
# Format: <tenant_uid_or_org_name> - must match the tenant your provider is authenticated
# against.
# terraform import spectrocloud_developer_setting.example my-org-name
import {
  to = spectrocloud_developer_setting.example
  id = "my-org-name"
}

# --- spectrocloud_platform_setting ---
# Format: <setting_id_or_name>:<project|tenant> - the id segment optionally carries a
# "platformsetting-" prefix (stripped automatically); name is then resolved to a UID in the
# given context.
# terraform import spectrocloud_platform_setting.example my-platform-setting:tenant
import {
  to = spectrocloud_platform_setting.example
  id = "my-platform-setting:tenant"
}

# --- spectrocloud_macros ---
# Format: <macros_id_or_name>:<project|tenant>
# terraform import spectrocloud_macros.example my-macros:project
import {
  to = spectrocloud_macros.example
  id = "my-macros:project"
}

# --- spectrocloud_sso ---
# Deviates from the common format: the second segment is the SSO protocol, not a project/tenant
# context - SSO is always tenant-scoped.
# Format: <tenant_uid_or_org_name>:<saml|oidc>
# terraform import spectrocloud_sso.example my-org-name:saml
import {
  to = spectrocloud_sso.example
  id = "my-org-name:saml"
}


##############################################################################
# Alerts & audit
##############################################################################

# --- spectrocloud_alert ---
# Deviates from the common format: the second segment is the alert component, not a
# project/tenant context. Currently the only supported component is "ClusterHealth".
# Format: <project_uid_or_name>:ClusterHealth
# terraform import spectrocloud_alert.example my-project:ClusterHealth
import {
  to = spectrocloud_alert.example
  id = "my-project:ClusterHealth"
}

# --- spectrocloud_audit_trail ---
# Format: <audit_trail_sink_uid> - always tenant-scoped, no context suffix. Currently only the
# Splunk sink type is supported.
# terraform import spectrocloud_audit_trail.example 5f6e7d8c9b0a1234567890cd
import {
  to = spectrocloud_audit_trail.example
  id = "5f6e7d8c9b0a1234567890cd"
}


##############################################################################
# Applications (Palette Dev Engine)
##############################################################################

# --- spectrocloud_application ---
# Format: <application_id_or_name> - the provider tries the "project" context first, then
# "tenant", so no explicit context suffix is needed (or accepted).
# terraform import spectrocloud_application.example my-app-deployment
import {
  to = spectrocloud_application.example
  id = "my-app-deployment"
}

# --- spectrocloud_application_profile ---
# Format: <profile_id_or_name>[:<project|tenant|system>[:<version>]] - context defaults to
# "project" and version defaults to "1.0.0" when omitted.
# terraform import spectrocloud_application_profile.example my-app-profile:project:1.0.0
import {
  to = spectrocloud_application_profile.example
  id = "my-app-profile:project:1.0.0"
}


##############################################################################
# KubeVirt (virtual machines on a host cluster)
##############################################################################
# Both resources below use a slash-separated composite ID (not the colon-separated
# `id:context` convention used elsewhere), since a VM/data volume is scoped to a specific
# Kubernetes cluster and namespace rather than a Palette project/tenant.

# --- spectrocloud_virtual_machine ---
# Format: <project|tenant>/<host_cluster_uid>/<namespace>/<vm_name>
# terraform import spectrocloud_virtual_machine.example project/64f1a2b3c4d5e6f7a8b9c0d1/default/my-vm
import {
  to = spectrocloud_virtual_machine.example
  id = "project/64f1a2b3c4d5e6f7a8b9c0d1/default/my-vm"
}

# --- spectrocloud_datavolume ---
# Format: <project|tenant>/<host_cluster_uid>/<vm_namespace>/<vm_name>/<datavolume_name>
# terraform import spectrocloud_datavolume.example project/64f1a2b3c4d5e6f7a8b9c0d1/default/my-vm/my-datavolume
import {
  to = spectrocloud_datavolume.example
  id = "project/64f1a2b3c4d5e6f7a8b9c0d1/default/my-vm/my-datavolume"
}
