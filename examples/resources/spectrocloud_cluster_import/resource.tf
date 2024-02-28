data "spectrocloud_cluster_profile" "profile" {
  # id = <uid>
  name = var.cluster_cluster_profile_name
}


resource "spectrocloud_cluster_import" "cluster" {
  name  = var.cluster_name
  cloud = var.cloud_type
  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }
  /*  pack {
       name   = "k8s-dashboard"
       tag    = "2.1.x"
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
    }*/
}
