# resource "spectrocloud_cluster_config_policy" "weekly_maintenance" {
#   name    = "weekly-maintenance-policy"
#   context = "project"

#   schedules {
#     name         = "sunday-maintenance"
#     start_cron   = "0 2 * * SUN"
#     duration_hrs = 4
#   }
# }

# Example with multiple schedules and tags
resource "spectrocloud_cluster_config_policy" "multi_schedule" {
  name    = "multi-schedule-policy-updated"
  context = "project"
  tags    = ["env:production", "team:devops", "test"]

  schedules {
    name         = "weekday-maintenance"
    start_cron   = "0 1 * * 1-5"
    duration_hrs = 2
  }

  schedules {
    name         = "weekend-maintenance"
    start_cron   = "1 3 * * 0,6"
    duration_hrs = 6
  }
}

# # Tenant-level maintenance policy
# resource "spectrocloud_cluster_config_policy" "tenant_policy" {
#   name    = "tenant-wide-maintenance"
#   context = "tenant"

#   schedules {
#     name         = "monthly-maintenance"
#     start_cron   = "0 3 1 * *"
#     duration_hrs = 8
#   }
# }

# Import example
# import {
#   to = spectrocloud_cluster_config_policy.imported_policy
#   id = "63d48062b3a0c92a6f230112"
# }

