resource "spectrocloud_cluster_config_template" "aws_template" {
  name       = "aws-prod-template"
  cloud_type = "aws"
  context    = "project"

  # Only one policy is supported (MaxItems: 1)
  # Policy can be replaced by changing the UID
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

    # variables {
    #   name            = "environment"
    #   value           = "production"
    #   assign_strategy = "cluster" # Cluster-specific override
    # }
  }

  # Trigger immediate cluster upgrade by setting/changing this timestamp
  # NOTE: This triggers NOW - it does NOT schedule a future upgrade
  # Set to current timestamp to trigger: upgrade_now = "2024-11-12T15:30:00Z"
}

# Example showing day 2 operations:
#
# 1. Triggering immediate cluster upgrades (uses PATCH /v1/spectroclusters/clusterTemplates/{uid}/clusters/upgrade):
#    # IMPORTANT: This triggers upgrade NOW, not at the specified time!
#    upgrade_now = "2024-01-15T10:30:00Z"  # Initial trigger (executes immediately)
#    
#    # To trigger another upgrade later, change to current timestamp:
#    upgrade_now = "2024-02-20T14:00:00Z"  # Second trigger (also executes immediately)
#
# 2. Replacing policy (uses PATCH /v1/clusterTemplates/{uid}/policies):
#    policies {
#      uid  = "new-policy-uid-123"  # Changed policy UID
#      kind = "maintenance"
#    }
#
# 3. Updating only variable values (uses PATCH /v1/clusterTemplates/{uid}/profiles/variables):
#    - Change "region" from "us-west-2" to "us-east-1"
#    - Change "instance_type" from "t3.medium" to "t3.large"
#
# 4. Adding/removing profiles (uses PUT /v1/clusterTemplates/{uid}/profiles):
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

