locals {
  falco_version = "1.16.3"
  prometheus_version = "30.0.3"
  fluentbit_version = "1.3.5"
}

data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "csi" {
  name    = "csi-azure"
  registry_uid = data.spectrocloud_registry.registry.id
  version = "1.0.0"
}

data "spectrocloud_pack" "cni" {
  name    = "cni-kubenet"
  registry_uid = data.spectrocloud_registry.registry.id
  version = "1.0.0"
}

data "spectrocloud_pack" "k8s" {
  name    = "kubernetes-aks"
  registry_uid = data.spectrocloud_registry.registry.id
  version = "1.23"
}

data "spectrocloud_pack" "ubuntu" {
  name    = "ubuntu-aks"
  registry_uid = data.spectrocloud_registry.registry.id
  version = "18.04"
}

data "spectrocloud_pack" "istio" {
  name    = "istio"
  registry_uid = data.spectrocloud_registry.registry.id
  version = "1.6.2"
}

data "spectrocloud_pack" "falco" {
  name    = "falco"
  registry_uid = data.spectrocloud_registry.registry.id
  version = local.falco_version
}

data "spectrocloud_pack" "prometheus-operator" {
  name    = "prometheus-operator"
  registry_uid = data.spectrocloud_registry.registry.id
  version = local.prometheus_version
}

data "spectrocloud_pack" "fluentbit" {
  name    = "fluentbit"
  registry_uid = data.spectrocloud_registry.registry.id
  version = local.fluentbit_version
}