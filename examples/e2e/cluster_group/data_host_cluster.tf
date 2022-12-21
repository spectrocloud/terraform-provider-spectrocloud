# Look up in data source for a host cluster.
data "spectrocloud_cluster" "host_cluster0" {
  name    = "eks-dev-nik-7-tenant0"
  context = "tenant"
}

data "spectrocloud_cluster" "host_cluster1" {
  name    = "eks-dev-nik-7-tenant1"
  context = "tenant"
}
