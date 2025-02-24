resource "spectrocloud_platform_setting" "platform_settings" {
  context                  = "tenant"
  enable_auto_remediation  = true
  session_timeout          = 230
  cluster_auto_remediation = true
  pause_agent_upgrades     = "lock"
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
