resource "spectrocloud_cluster_config_template" "aws_template" {
  name       = "aws-prod-template"
  cloud_type = "aws"

  profiles {
    uid = var.cluster_profile_infra_id
  }

  profiles {
    uid = var.cluster_profile_addon_id
  }

  policies {
    uid  = var.maintenance_policy_id
    kind = "maintenance"
  }
}

# Minimal example
resource "spectrocloud_cluster_config_template" "minimal" {
  name       = "minimal-template"
  cloud_type = "azure"
}

# Import example
# import {
#   to = spectrocloud_cluster_config_template.imported_template
#   id = "63d48062b3a0c92a6f230112"
# }

