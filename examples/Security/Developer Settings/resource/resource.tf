# Manages tenant-level developer quota settings (used by developer/virtual clusters). This is a
# singleton per tenant - there is nothing to name or ForceNew here; every attribute updates in
# place.
#
# Attributes:
#   virtual_clusters_limit    - Optional, default 2. Number of virtual clusters a developer can
#                                create. Allowed range: 1-1000.
#   cpu                       - Optional, default 12. CPU cores allocated to the developer
#                                cluster group. Allowed range: 4-1000.
#   memory                    - Optional, default 16. Memory in GiB allocated to the developer
#                                cluster group. Allowed range: 4-1000.
#   storage                   - Optional, default 20. Storage in GiB allocated to the developer
#                                cluster group. Allowed range: 2-100000.
#   hide_system_cluster_group - Optional, default false. If true, hides the system cluster group
#                                from developers.
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