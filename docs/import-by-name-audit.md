# Terraform Provider Spectrocloud – Import by Name Audit

**Date:** February 2025  
**Scope:** All resources in `terraform-provider-spectrocloud` that support Terraform import.

---

## Summary

Across the provider, **only 2 resources** support import by name today. **All other resources with a custom Importer** accept only UID/ID (or a structured ID such as `context:uid`) and need import-by-name added for consistency.

---

## 1. Resources That Already Support Import by Name

| Resource Type | Implementation Notes |
|---------------|------------------------|
| **spectrocloud_registry_oci** | Tries ECR/Basic by UID first, then falls back to `GetOciRegistryByName(importID)` when UID lookup fails. Import ID can be UID or registry name. |
| **spectrocloud_alert** | Import ID format: `projectIdentifier:component`. Project part can be **project name**: uses `GetProject(projectIdentifier)` then `GetProjectUID(projectIdentifier)` on failure. Component must be `ClusterHealth`. |

---

## 2. Resources That Need Import by Name Implemented

These resources have an Importer but resolve only by UID/ID (or `context:uid`). They should be extended to accept name where the API supports it (e.g. `GetXByName` in palette-sdk-go or equivalent).

### 2.1 Core / Organization

| Resource Type | Current Behavior |
|---------------|------------------|
| spectrocloud_project | GetProject(projectUID) only |
| spectrocloud_team | GetTeam(teamUID) only |
| spectrocloud_user | GetUserByID(d.Id()) only |
| spectrocloud_role | GetRoleByID(d.Id()) only |
| spectrocloud_workspace | GetWorkspace(workspaceUID) only |
| spectrocloud_sso | Custom ID parsing, no name resolution |

### 2.2 Cluster Profiles and Config

| Resource Type | Current Behavior |
|---------------|------------------|
| spectrocloud_cluster_profile | ParseResourceID + GetClusterProfile(profileID) only |
| spectrocloud_cluster_config_template | Import by UID only |
| spectrocloud_cluster_config_policy | Import by UID only |

### 2.3 Clusters (ID format: clusterID:context)

| Resource Type |
|---------------|
| spectrocloud_cluster_aws |
| spectrocloud_cluster_azure |
| spectrocloud_cluster_aks |
| spectrocloud_cluster_eks |
| spectrocloud_cluster_gcp |
| spectrocloud_cluster_gke |
| spectrocloud_cluster_vsphere |
| spectrocloud_cluster_maas |
| spectrocloud_cluster_openstack |
| spectrocloud_cluster_apache_cloudstack |
| spectrocloud_cluster_custom_cloud |
| spectrocloud_cluster_edge_native |
| spectrocloud_cluster_edge_vsphere |
| spectrocloud_virtual_cluster |
| spectrocloud_cluster_brownfield |
| spectrocloud_cluster_group |

### 2.4 Cloud Accounts (ID format: context:accountID)

| Resource Type |
|---------------|
| spectrocloud_cloudaccount_aws |
| spectrocloud_cloudaccount_azure |
| spectrocloud_cloudaccount_gcp |
| spectrocloud_cloudaccount_vsphere |
| spectrocloud_cloudaccount_openstack |
| spectrocloud_cloudaccount_apache_cloudstack |
| spectrocloud_cloudaccount_maas |
| spectrocloud_cloudaccount_custom |

### 2.5 Registries

| Resource Type | Current Behavior |
|---------------|------------------|
| spectrocloud_registry_helm | GetHelmRegistry(registryUID) only. SDK has GetHelmRegistryByName; can be used for import-by-name. |

### 2.6 Applications and Appliance

| Resource Type | Current Behavior |
|---------------|------------------|
| spectrocloud_application | GetApplication(applicationID) only |
| spectrocloud_application_profile | UID only |
| spectrocloud_appliance | GetAppliance(applianceUID) only. Data source uses GetApplianceByName. |

### 2.7 Backup and Infrastructure

| Resource Type | Current Behavior |
|---------------|------------------|
| spectrocloud_backup_storage_location | ID format bsl_id or context:bsl_id; UID only |
| spectrocloud_privatecloudgateway_ippool | UID only |
| spectrocloud_privatecloudgateway_dns_map | UID only |

### 2.8 IAM, Settings, Filters

| Resource Type | Current Behavior |
|---------------|------------------|
| spectrocloud_filter | UID only. Data source has GetTagFilterByName. |
| spectrocloud_macros | UID only |
| spectrocloud_ssh_key | ID format ssh_key_id or context:ssh_key_id; UID only. Data source has GetSSHKeyByName. |
| spectrocloud_password_policy | UID only |
| spectrocloud_resource_limit | UID only |
| spectrocloud_developer_setting | UID only |
| spectrocloud_platform_setting | UID only |
| spectrocloud_registration_token | UID only. Data source has GetRegistrationTokenByName. |

---

## 3. Resources With Pass-Through or Special Import (Optional / Later)

| Resource Type | Notes |
|---------------|--------|
| spectrocloud_virtual_machine (Kubevirt) | Uses ImportStatePassthroughContext; composite IDs. Import by name would require defining how name maps to composite ID (e.g. cluster + namespace + name). |
| spectrocloud_datavolume | Same as above. |
| spectrocloud_cluster_profile_import | Import feature resource; create/read/update/delete semantics, not classic import-by-name. |

---

## 4. Implementation Pattern for Adding Import by Name

1. **Try UID first:** Treat import ID as UID and call the existing Get-by-UID API.
2. **Fallback to name:** If that fails (e.g. 404) or if the ID looks like a name (e.g. no UUID format, contains spaces), call the ByName API if it exists in the SDK (e.g. GetHelmRegistryByName, GetWorkspaceByName, GetClusterByName).
3. **Set state:** Set `d.SetId(uid)` and any required attributes, then run the resource Read to populate state.

Where no ByName API exists, add it in palette-sdk-go or document that the resource remains UID-only until the API supports name lookup.

Reference implementation: `spectrocloud/resource_registry_oci_import.go`.

---

## 5. Generating a PDF From This Document

- **Cursor/VS Code:** Use a "Markdown: Export to PDF" extension.
- **Command line:** `pandoc docs/import-by-name-audit.md -o import-by-name-audit.pdf`
- **Browser:** Open the `.md` file in a viewer that supports Markdown, then use Print → Save as PDF.
