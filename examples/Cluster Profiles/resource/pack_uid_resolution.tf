# Pack UID resolution: alternatives to resource.tf's explicit `uid` on each pack. Same resource
# type and Day-2 mutability as resource.tf in this folder - see that file's top comment.
#
# `uid` is Optional/Computed on every pack - when omitted, the provider resolves it internally
# from name + tag + registry_uid (or registry_name) instead. Two packs are enough to show this;
# it works identically for however many packs a profile has.

data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
}

resource "spectrocloud_cluster_profile" "profile_with_auto_resolution" {
  name        = "example-profile-auto-resolution"
  description = "Demonstrates automatic pack UID resolution"
  tags        = ["example", "auto-resolution"]
  cloud       = "aws"
  type        = "cluster"
  version     = "1.0.0"

  # Operating System pack - uid resolved automatically from name+tag+registry_uid
  pack {
    name         = "ubuntu-aws"
    tag          = "22.04"
    registry_uid = data.spectrocloud_registry.public_registry.id
    values       = <<-EOT
      timezone: UTC
      package_update: true
    EOT
  }

  # Kubernetes pack - same resolution
  pack {
    name         = "kubernetes"
    tag          = "1.27.5"
    registry_uid = data.spectrocloud_registry.public_registry.id
    values       = <<-EOT
      kubeadmconfig:
        apiServer:
          extraArgs:
            audit-log-maxage: "30"
            audit-log-maxbackup: "10"
        kubernetesVersion: "v1.27.5"
    EOT
  }
}

# Example mixing automatic resolution with explicit UIDs
resource "spectrocloud_cluster_profile" "profile_mixed_approach" {
  name        = "example-profile-mixed"
  description = "Demonstrates mixing automatic resolution with explicit UIDs"
  tags        = ["example", "mixed-approach"]
  cloud       = "aws"
  type        = "cluster"
  version     = "1.0.0"

  # Pack with explicit UID (traditional approach)
  pack {
    name   = "ubuntu-aws"
    tag    = "22.04"
    uid    = "example-explicit-uid-1234"
    values = "timezone: UTC"
  }

  # Pack with automatic resolution (new approach)
  pack {
    name         = "kubernetes"
    tag          = "1.27.5"
    registry_uid = data.spectrocloud_registry.public_registry.id
    values       = <<-EOT
      kubeadmconfig:
        kubernetesVersion: "v1.27.5"
    EOT
  }
}

# Output the created cluster profile IDs
output "auto_resolution_profile_id" {
  description = "ID of the cluster profile created with automatic pack UID resolution"
  value       = spectrocloud_cluster_profile.profile_with_auto_resolution.id
}

output "mixed_approach_profile_id" {
  description = "ID of the cluster profile created with mixed approach"
  value       = spectrocloud_cluster_profile.profile_mixed_approach.id
} 