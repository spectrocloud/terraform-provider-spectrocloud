data "spectrocloud_cluster_group" "cluster-group" {
  name    = var.cluster-group-name
  context = "project"
}

data "spectrocloud_registry" "public_registry" {
  name = "Public Repo"
}

data "spectrocloud_pack_simple" "container_pack" {
  type         = "container"
  name         = "container"
  version      = "1.0.2"
  registry_uid = data.spectrocloud_registry.public_registry.id
}

data "spectrocloud_pack_simple" "postgres_service" {
  name         = "postgresql-operator"
  type         = "operator-instance"
  version      = "1.8.2"
  registry_uid = data.spectrocloud_registry.public_registry.id
}
