resource "spectrocloud_cluster_config_template" "aws_template" {
  name       = "aws-prod-template"
  cloud_type = "aws"
  context    = "project"

  policies {
    uid  = var.maintenance_policy_id
    kind = "maintenance"
  }

  profiles {
    uid = var.addon_profile_id
  }

  profiles {
    uid = "69130518a2d75382d3f0ee89"
  }

 
}

# Minimal example
# resource "spectrocloud_cluster_config_template" "minimal" {
#   name       = "minimal-template"
#   cloud_type = "azure"
# }

# Import example
# import {
#   to = spectrocloud_cluster_config_template.imported_template
#   id = "63d48062b3a0c92a6f230112"
# }

