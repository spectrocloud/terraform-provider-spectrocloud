resource "spectrocloud_cluster_profile" "addon_profile" {
  name        = "addon-demo"
  description = "Addon profile for demo"
  #tags        = ["owner:bob"]
  cloud       = "all"
  type        = "add-on"

  pack {
    name   = var.argocd_name
    tag    = var.argocd_version
    uid    = data.spectrocloud_pack.argo-cd.id
    values = file("config/argocd.yaml")
  }
}