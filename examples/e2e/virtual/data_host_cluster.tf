# Look up in data source for a host cluster.
data "spectrocloud_cluster" "host_cluster" {
  name = "test-new-eks-host"
}
