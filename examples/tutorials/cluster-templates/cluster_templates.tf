resource "spectrocloud_cluster_config_template" "aws_template" {
  count = var.deploy-aws ? 1 : 0

  name       = "tf-cluster-template-aws"
  cloud_type = "aws"
  context    = "project"

  cluster_profile {
    id = (var.create_new_profile_version && var.update_template_profile_version) ? spectrocloud_cluster_profile.aws_profile_v110[0].id : spectrocloud_cluster_profile.aws_profile[0].id
  }

  policy {
    id   = spectrocloud_cluster_config_policy.maintenance.id
    kind = "maintenance"
  }

  upgrade_now = var.upgrade_now_timestamp != "" ? var.upgrade_now_timestamp : null
}

resource "spectrocloud_cluster_config_template" "azure_template" {
  count = var.deploy-azure ? 1 : 0

  name       = "tf-cluster-template-azure"
  cloud_type = "azure"
  context    = "project"

  cluster_profile {
    id = (var.create_new_profile_version && var.update_template_profile_version) ? spectrocloud_cluster_profile.azure_profile_v110[0].id : spectrocloud_cluster_profile.azure_profile[0].id
  }

  policy {
    id   = spectrocloud_cluster_config_policy.maintenance.id
    kind = "maintenance"
  }

  upgrade_now = var.upgrade_now_timestamp != "" ? var.upgrade_now_timestamp : null
}
