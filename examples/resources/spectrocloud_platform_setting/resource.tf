# Manages tenant- or project-scoped platform settings (singleton per context). Day-2 mutability:
# only `context` is ForceNew - changing it recreates the resource. Everything else updates in
# place, subject to the context restrictions noted below.
#
# When context = "project", the provider rejects session_timeout, login_banner,
# non_fips_addon_pack, non_fips_features, and non_fips_cluster_import (all tenant-only) - set
# context = "tenant" (as below) to use them.
resource "spectrocloud_platform_setting" "platform_settings" {
  # Optional, default "tenant", ForceNew. Allowed: "project", "tenant".
  context = "tenant"

  # Optional, default true. Despite the schema description ("project context only"), this is
  # accepted and takes effect under "tenant" context too, as used here.
  enable_auto_remediation = true

  # Optional. Minutes of inactivity before auto logout (tenant context only). Default in Palette
  # is 240 when unset.
  session_timeout = 230

  # Optional, default true. Auto-replaces unhealthy nodes on Palette-provisioned clusters. Not
  # applicable to EKS, AKS, or TKE clusters.
  cluster_auto_remediation = false

  # Optional, default false. Enables automatic cluster role binding.
  automatic_cluster_role_binding = false

  # Optional. Vertex-only: allow non-FIPS-compliant addon packs (tenant context only).
  non_fips_addon_pack = true

  # Optional. Vertex-only: allow non-FIPS-compliant features like backup/restore/scans (tenant
  # context only).
  non_fips_features = true

  # Optional. Vertex-only: allow importing clusters that may not be FIPS-compliant (tenant
  # context only).
  non_fips_cluster_import = true

  # Optional, default "unlock". "lock" pauses automatic Palette agent/component upgrades for
  # clusters under this tenant/project; "unlock" allows them.
  pause_agent_upgrades = "lock"

  # Optional, at most one block (tenant context only). Banner users must acknowledge before
  # signing in.
  login_banner {
    title   = "test"
    message = "test"
  }
}

## import existing platform settings
#import {
#  to = spectrocloud_platform_setting.platform_setting
#  id = "{tenantUID/ProjectUID}:{tenant/project)}" // tenant-uid:tenant or project-uid:project
#}
