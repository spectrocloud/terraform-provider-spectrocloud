resource "spectrocloud_cluster_config_template" "aws_template" {
  name       = "aws-prod-template"
  cloud_type = "aws"
  context    = "project"

  policies {
    uid  = "69131adb05561b51307764e5"
    kind = "maintenance"
  }

  profiles {
    uid = var.addon_profile_id
    
    # # Profile variables with assignment strategies
    # variables {
    #   name            = "region"
    #   value           = "us-west-2"
    #   assign_strategy = "all" # Apply to all clusters
    # }
    
    # variables {
    #   name            = "instance_type"
    #   value           = "t3.medium"
    #   assign_strategy = "all"
    # }
  }

  profiles {
    uid = "69130518a2d75382d3f0ee89"
    
    variables {
      name            = "environment"
      value           = "production"
      assign_strategy = "cluster" # Cluster-specific override
    }
  }
}

# Example showing day 2 operations:
# 1. Updating only variable values (uses PATCH endpoint):
#    - Change "region" from "us-west-2" to "us-east-1"
#    - Change "instance_type" from "t3.medium" to "t3.large"
#
# 2. Adding/removing profiles (uses PUT endpoint):
#    - Add a new profile block with different UID
#    - Remove an existing profile block
#
# Terraform will automatically detect which type of change occurred
# and use the appropriate API endpoint.

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

