# Day-2 mutability: only `cloud_type` is ForceNew - changing it recreates the template. `name`,
# `context`, `policy`, `cluster_profile` (including its variable overrides), and `upgrade_now`
# all update in place. `execution_state` and `attached_cluster` are Computed/read-only - Palette
# populates them, they cannot be set here.
resource "spectrocloud_cluster_config_template" "aws_template" {
  name       = "aws-prod-template"
  cloud_type = "aws"
  context    = "project"

  # Only one policy is supported (MaxItems: 1)
  # Policy can be replaced by changing the ID
  policy {
    id   = var.maintenance_policy_id
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
    id = var.infra_profile_id

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
# Import an existing cluster config template. The ID must be
# "<template_id_or_name>:<project|tenant>" - the context suffix is required, not optional;
# omitting it causes the import to fail.
#
# import {
#   to = spectrocloud_cluster_config_template.imported_template
#   id = "63d48062b3a0c92a6f230112:project"
# }
