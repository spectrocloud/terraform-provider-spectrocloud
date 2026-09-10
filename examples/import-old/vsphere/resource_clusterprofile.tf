# If looking up a cluster profile instead of creating a new one
# data "spectrocloud_cluster_profile" "profile" {
#   # id = <uid>
#   name = var.cluster_cluster_profile_name
# }

data "spectrocloud_pack" "k8s_dashboard" {
  name    = "k8s-dashboard"
  version = "2.1.0"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "cloud-addon-12"
  description = "add-on cp"
  cloud       = "vsphere"
  type        = "add-on"

  pack {
    name   = "k8s-dashboard"
    tag    = "2.1.x"
    uid    = data.spectrocloud_pack.k8s_dashboard.id
    values = <<-EOT
      manifests:
        k8s-dashboard:

          #Namespace to install kubernetes-dashboard
          namespace: "kubernetes-dashboard"

          #The ClusterRole to assign for kubernetes-dashboard. By default, a ready-only cluster role is provisioned
          clusterRole: "k8s-dashboard-readonly"

          #Self-Signed Certificate duration in hours
          certDuration: 9000h

          #Self-Signed Certificate renewal in hours
          certRenewal: 720h     #30d

          #The service type for dashboard. Supported values are ClusterIP / LoadBalancer / NodePort
          serviceType: ClusterIP

          #Flag to enable skip login option on the dashboard login page
          skipLogin: false

          #Ingress config
          ingress:
            enabled: false
    EOT
  }
}
