# Manages tenant- or project-scoped platform settings (singleton per context). Day-2 mutability:
# only `context` is ForceNew - changing it recreates the resource. Everything else updates in
# place, subject to the context restrictions noted below.
#
# When context = "project", the provider rejects session_timeout, login_banner,
# non_fips_addon_pack, non_fips_features, and non_fips_cluster_import (all tenant-only) - set
# context = "tenant" (as below) to use them.
#
# Attributes:
#   context                        - Optional, default "tenant", ForceNew. Allowed: "project",
#                                     "tenant".
#   enable_auto_remediation       - Optional, default true. Despite the schema description
#                                     ("project context only"), this is accepted and takes
#                                     effect under "tenant" context too, as used here.
#   session_timeout                - Optional. Minutes of inactivity before auto logout (tenant
#                                     context only). Default in Palette is 240 when unset.
#   cluster_auto_remediation       - Optional, default true. Auto-replaces unhealthy nodes on
#                                     Palette-provisioned clusters. Not applicable to EKS, AKS,
#                                     or TKE clusters.
#   automatic_cluster_role_binding - Optional, default false. Enables automatic cluster role
#                                     binding.
#   non_fips_addon_pack            - Optional. Vertex-only: allow non-FIPS-compliant addon packs
#                                     (tenant context only).
#   non_fips_features              - Optional. Vertex-only: allow non-FIPS-compliant features
#                                     like backup/restore/scans (tenant context only).
#   non_fips_cluster_import        - Optional. Vertex-only: allow importing clusters that may
#                                     not be FIPS-compliant (tenant context only).
#   pause_agent_upgrades           - Optional, default "unlock". "lock" pauses automatic
#                                     Palette agent/component upgrades for clusters under this
#                                     tenant/project; "unlock" allows them.
#
# login_banner (optional, at most one block, tenant context only): banner users must acknowledge
# before signing in.
#   title   - Banner title.
#   message - Banner message.
resource "spectrocloud_platform_setting" "platform_settings" {
  context                        = "tenant"
  enable_auto_remediation        = true
  session_timeout                = 230
  cluster_auto_remediation       = false
  automatic_cluster_role_binding = false
  non_fips_addon_pack            = true
  non_fips_features              = true
  non_fips_cluster_import        = true
  pause_agent_upgrades           = "lock"

  login_banner {
    title   = "test"
    message = "test"
  }
}

# import existing platform settings
#import {
#  to = spectrocloud_platform_setting.platform_setting
#  id = "{tenantUID/ProjectUID}:{tenant/project)}" // tenant-uid:tenant or project-uid:project
#}
