# Manages tenant-level developer quota settings (used by developer/virtual clusters). This is a
# singleton per tenant - there is nothing to name or ForceNew here; every attribute updates in
# place.
resource "spectrocloud_developer_setting" "dev_setting" {
  # Optional, default 2. Number of virtual clusters a developer can create. Allowed range: 1-1000.
  virtual_clusters_limit = 10

  # Optional, default 12. CPU cores allocated to the developer cluster group. Allowed range: 4-1000.
  cpu = 20

  # Optional, default 16. Memory in GiB allocated to the developer cluster group. Allowed range: 4-1000.
  memory = 100

  # Optional, default 20. Storage in GiB allocated to the developer cluster group. Allowed range: 2-100000.
  storage = 100

  # Optional, default false. If true, hides the system cluster group from developers.
  hide_system_cluster_group = false
}

## import existing developer settings
#import {
#  to = spectrocloud_developer_setting.dev_setting
#  id = "{tenantUID}" // tenant-uid
#}