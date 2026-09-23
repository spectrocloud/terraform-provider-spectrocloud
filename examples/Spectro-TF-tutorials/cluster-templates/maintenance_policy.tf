resource "spectrocloud_cluster_config_policy" "maintenance" {
  name    = "tf-maintenance-policy"
  context = "project"

  schedules {
    name         = "weekly-sunday"
    start_cron   = "0 0 * * SUN"
    duration_hrs = 4
  }
}
