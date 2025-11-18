resource "spectrocloud_cluster_config_template" "aws_template" {
  name        = "aks-tf-si"
  cloud_type  = "aws"
  context     = "tenant"
  upgrade_now = "2024-11-13T15:30:00Z"

  # Only one policy is supported (MaxItems: 1)
  # Policy can be replaced by changing the UID
  policies {
    uid  = spectrocloud_cluster_config_policy.weekly_maintenance.id
    kind = "maintenance"
  }

  profiles {
    uid = "691b556e617bd79e8c6de03a" # spectrocloud_cluster_profile.infra_profile.id
  }

  profiles {
    uid = "691b556e50498bf5109ecf19" # spectrocloud_cluster_profile.addon_profile1.id
  }

  profiles {
    uid = "691b556e50498bf514992e1f" # spectrocloud_cluster_profile.addon_profile2.id
  }

  # Optional: Trigger immediate cluster upgrade for all attached clusters
  # NOTE: This triggers upgrade NOW - it does NOT schedule a future upgrade
  # Uncomment and set to current timestamp when you want to trigger an upgrade:
  # upgrade_now = "2024-11-12T15:30:00Z"
}
