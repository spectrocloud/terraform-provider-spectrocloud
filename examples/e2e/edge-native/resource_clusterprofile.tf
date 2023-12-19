#Define a public repo as registry for packs
data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "cni" {
  registry_uid = data.spectrocloud_registry.registry.id
  name         = "cni-calico"
  version      = "3.26.1"
}

data "spectrocloud_pack" "k8s" {
  registry_uid = data.spectrocloud_registry.registry.id
  name         = "edge-k3s"
  version      = "1.27.2"
}

data "spectrocloud_pack" "os" {
  registry_uid = data.spectrocloud_registry.registry.id
  name         = "edge-native-byoi"
  version      = "1.0.0"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "edge-profile-tf"
  description = "basic cp"
  tags        = ["dev", "department:devops", "owner:alice"]
  cloud       = "edge-native"
  type        = "cluster"

  pack {
    name = data.spectrocloud_pack.os.name
    tag  = data.spectrocloud_pack.os.version
    uid  = data.spectrocloud_pack.os.id
    #values = data.spectrocloud_pack.os.values
    values = <<-EOT
      pack:
        content:
          images:
            - image: "{{.spectro.pack.edge-native-byoi.options.system.uri}}"
      options:
        system.uri: "{{ .spectro.pack.edge-native-byoi.options.system.registry }}/{{ .spectro.pack.edge-native-byoi.options.system.repo }}:{{ .spectro.pack.edge-native-byoi.options.system.k8sDistribution }}-{{ .spectro.system.kubernetes.version }}-{{ .spectro.pack.edge-native-byoi.options.system.peVersion }}-{{ .spectro.pack.edge-native-byoi.options.system.customTag }}"
        system.registry: harbor.mycompany.tld
        system.repo: ubuntu
        system.k8sDistribution: k3s
        system.osName: ubuntu
        system.peVersion: v4.1.2
        system.customTag: mytag
        system.osVersion: 22
    EOT
  }

  pack {
    name   = data.spectrocloud_pack.k8s.name
    tag    = data.spectrocloud_pack.k8s.version
    uid    = data.spectrocloud_pack.k8s.id
    values = data.spectrocloud_pack.k8s.values
  }

  pack {
    name   = data.spectrocloud_pack.cni.name
    tag    = data.spectrocloud_pack.cni.version
    uid    = data.spectrocloud_pack.cni.id
    values = data.spectrocloud_pack.cni.values
  }
}
