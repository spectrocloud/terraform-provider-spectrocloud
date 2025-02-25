resource "spectrocloud_developer_setting" "dev_setting" {
  virtual_clusters_limit    = 10
  cpu                       = 20
  memory                    = 100
  storage                   = 100
  hide_system_cluster_group = false
}

## import existing developer settings
#import {
#  to = spectrocloud_developer_setting.dev_setting
#  id = "{tenantUID}" // tenant-uid
#}