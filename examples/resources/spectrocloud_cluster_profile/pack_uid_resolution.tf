# Example demonstrating automatic pack UID resolution
# This example shows how to create a cluster profile without explicitly 
# specifying pack UIDs - the system will resolve them automatically using
# the pack name, tag, and registry_uid.

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

  # Operating System pack - automatically resolved
  pack {
    name         = "ubuntu-aws"
    tag          = "22.04"
    registry_uid = data.spectrocloud_registry.public_registry.id
    values       = <<-EOT
      timezone: UTC
      package_update: true
    EOT
  }

  # Kubernetes pack - automatically resolved
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

  # CNI pack - automatically resolved
  pack {
    name         = "cni-calico"
    tag          = "3.26.1"
    registry_uid = data.spectrocloud_registry.public_registry.id
    values       = <<-EOT
      manifests:
        calico:
          contents: |
            # Calico configuration
            apiVersion: operator.tigera.io/v1
            kind: Installation
            metadata:
              name: default
            spec:
              calicoNetwork:
                ipPools:
                - blockSize: 26
                  cidr: 10.244.0.0/16
                  encapsulation: VXLANCrossSubnet
    EOT
  }

  # CSI pack - automatically resolved
  pack {
    name         = "csi-aws-ebs"
    tag          = "1.22.0"
    registry_uid = data.spectrocloud_registry.public_registry.id
    values       = <<-EOT
      manifests:
        ebs-csi-driver:
          contents: |
            # AWS EBS CSI Driver configuration
    EOT
  }

  # Manifest pack - no UID resolution needed for manifest type
  pack {
    name = "custom-manifests"
    type = "manifest"
    tag  = "1.0.0"
    manifest {
      name    = "example-namespace"
      content = <<-EOT
        apiVersion: v1
        kind: Namespace
        metadata:
          name: example-app
          labels:
            purpose: example
      EOT
    }
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