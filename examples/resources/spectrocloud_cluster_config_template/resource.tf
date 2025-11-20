resource "spectrocloud_cluster_config_template" "aws_template" {
  name       = "aws-prod-template"
  cloud_type = "aws"
  context    = "project"

  # Only one policy is supported (MaxItems: 1)
  # Policy can be replaced by changing the ID
  policy {
    id   = "69131adb05561b51307764e5"
    kind = "maintenance"
  }

  cluster_profile {
    id = var.addon_profile_id

    # Profile variables with assignment strategies
    variables {
      name            = "region"
      value           = "us-west-2"
      assign_strategy = "all" # Apply to all clusters
    }

    variables {
      name            = "instance_type"
      value           = "t3.medium"
      assign_strategy = "all"
    }
  }

  cluster_profile {
    id = "69130518a2d75382d3f0ee89"

    variables {
      name            = "environment"
      value           = "production"
      assign_strategy = "cluster" # Cluster-specific override
    }
  }

  # Optional: Trigger immediate cluster upgrade for all attached clusters
  # NOTE: This triggers upgrade NOW - it does NOT schedule a future upgrade
  # Uncomment and set to current timestamp when you want to trigger an upgrade:
  # upgrade_now = "2024-11-12T15:30:00Z"
}


# ═══════════════════════════════════════════════════════════════════════════
# IMPORT EXAMPLE
# ═══════════════════════════════════════════════════════════════════════════
# Import an existing cluster config template using its UID
#
# import {
#   to = spectrocloud_cluster_config_template.imported_template
#   id = "63d48062b3a0c92a6f230112"
# }
