resource "spectrocloud_cluster_profile" "ck8s_oidc" {
  name    = "ck8s-oidc"
  version = "1.0.0"
  type    = "cluster"
  context = "project"

  pack {
    name   = "ck8s"
    type   = "spectro"
    tag    = "1.32.0"
    uid    = data.spectrocloud_pack.ck8s.id
    values = <<-EOT
      pack:
        k8sHardening: false
      kubernetesVersion: "1.32.0"
      clusterConfig:
        controlPlane:
          extraKubeAPIServerArgs:
            --oidc-issuer-url: "{{.spectro.system.oidc.issuerurl}}"
            --oidc-client-id: "{{.spectro.system.oidc.clientid}}"
            --oidc-username-claim: "{{.spectro.system.oidc.usernameclaim}}"
            --oidc-groups-claim: "{{.spectro.system.oidc.groupsclaim}}"
    EOT
  }
}
