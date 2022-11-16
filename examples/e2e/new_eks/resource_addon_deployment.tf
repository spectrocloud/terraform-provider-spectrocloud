resource "spectrocloud_addon_deployment" "depl" {
  cluster_uid = spectrocloud_cluster_eks.cluster.id

  cluster_profile {
    id = spectrocloud_cluster_profile.profile_resource.id
    pack {
      name   = "rook-orchestrator"
      type   = "manifest"
      values = <<-EOT
      pack:
        spectrocloud.com/install-priority: "25"
    EOT
      manifest {
        name = "enable-rook-orchestrator"
        content = templatefile(local.templates["enable-rook-orchestrator_config"].location,
          {
            rook_ceph_version = "v1.9.2"
        })
      }
    }
  }
}
