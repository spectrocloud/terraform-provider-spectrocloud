locals {
  proxy_val = <<-EOT
        manifests:
          spectro-proxy:
            namespace: "cluster-{{ .spectro.system.cluster.uid }}"

            server: "{{ .spectro.system.reverseproxy.server }}"

            # Cluster UID - DO NOT CHANGE (new3)
            clusterUid: "{{ .spectro.system.cluster.uid }}"
            subdomain: "cluster-{{ .spectro.system.cluster.uid }}"
  EOT
}

resource "spectrocloud_cluster_profile" "infra_profile" {
  name        = "aks-infra"
  description = "Infra Cluster Profile"
  tags        = ["owner:dmitry"]
  cloud       = "aks"
  type        = "cluster"

  pack {
    name   = data.spectrocloud_pack.ubuntu.name
    tag    = "LTS__18.4.x"
    uid    = data.spectrocloud_pack.ubuntu.id
    values = "foo: 1"
  }

  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = "1.23.5"
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = "1.0.0"
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }

  pack {
    name   = data.spectrocloud_pack.csi.name
    tag    = "1.0.0"
    uid    = data.spectrocloud_pack.csi.id
    values = data.spectrocloud_pack.csi.values
  }
}

resource "spectrocloud_cluster_profile" "addon_profile" {
  name        = "application-standard"
  description = "Application Cluster Profile"
  tags        = ["owner:dmitry"]
  cloud       = "all"
  type        = "add-on"

  pack {
    name   = data.spectrocloud_pack.istio.name
    tag    = "1.6.2"
    uid    = data.spectrocloud_pack.istio.id
    values = data.spectrocloud_pack.istio.values
  }

  pack {
    name   = data.spectrocloud_pack.falco.name
    tag    = local.falco_version
    uid    = data.spectrocloud_pack.falco.id
    values = data.spectrocloud_pack.falco.values
    registry_uid = data.spectrocloud_registry.registry.id
  }

 pack {
   name   = data.spectrocloud_pack.prometheus-operator.name
   tag    = local.prometheus_version
   uid    = data.spectrocloud_pack.prometheus-operator.id
   values = replace(data.spectrocloud_pack.prometheus-operator.values, "adminPassword: ", "adminPassword: admin-pr0mEtheus")
 }

 pack {
    name   = data.spectrocloud_pack.fluentbit.name
    tag    = local.fluentbit_version
    uid    = data.spectrocloud_pack.fluentbit.id
    values = data.spectrocloud_pack.fluentbit.values
 }

  pack {
    name = "manifest-namespace"
    type = "manifest"
    manifest {
      name    = "manifest-namespace"
      content = <<-EOT
        apiVersion: v1
        kind: Namespace
        metadata:
          labels:
            app: wordpress
            app3: wordpress786
          name: wordpress
      EOT
    }
  }
}
