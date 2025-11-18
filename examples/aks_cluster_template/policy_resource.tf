resource "spectrocloud_cluster_config_policy" "weekly_maintenance" {
  name    = "ran-tf-si"
  context = "tenant"
  tags    = ["QA:Ranjith", "Dev:Siva"]

  schedules {
    name         = "sunday-maintenance"
    start_cron   = "0 4 * * 0"
    duration_hrs = 2
  }

  #   schedules {
  #    name         = "weekday-maintenance"
  #    start_cron   = "0 1 * * 1-5"
  #    duration_hrs = 2
  #  }
}