resource "spectrocloud_cluster_config_policy" "weekly_maintenance" {
  name = "weekly-maintenance-policy"

  schedules {
    name         = "sunday-maintenance"
    start_cron   = "0 2 * * SUN"
    duration_hrs = 4
  }
}

# Example with multiple schedules
resource "spectrocloud_cluster_config_policy" "multi_schedule" {
  name = "multi-schedule-policy"

  schedules {
    name         = "weekday-maintenance"
    start_cron   = "0 1 * * 1-5"
    duration_hrs = 2
  }

  schedules {
    name         = "weekend-maintenance"
    start_cron   = "0 3 * * 0,6"
    duration_hrs = 6
  }
}

# Import example
# import {
#   to = spectrocloud_cluster_config_policy.imported_policy
#   id = "63d48062b3a0c92a6f230112"
# }

