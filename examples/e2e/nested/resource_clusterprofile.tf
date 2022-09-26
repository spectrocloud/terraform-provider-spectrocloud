# nested cluster can only have an addon profile.
data "spectrocloud_registry" "registry" {
  name = "Public Repo"
}

data "spectrocloud_pack" "elastic-fluentd-kibana" {
  name = "elastic-fluentd-kibana"
  registry_uid = data.spectrocloud_registry.registry.id
  version  = "6.7.0"
}

resource "spectrocloud_cluster_profile" "profile" {
  name        = "elastic search"
  description = "elastic search addon"
  cloud       = "all"
  type        = "add-on"


  pack {
    name   = data.spectrocloud_pack.elastic-fluentd-kibana.name
    tag    = data.spectrocloud_pack.elastic-fluentd-kibana.version
    uid    = data.spectrocloud_pack.elastic-fluentd-kibana.id
    values = data.spectrocloud_pack.elastic-fluentd-kibana.values
  }
}
