
resource "spectrocloud_cluster_openstack" "cluster" {
  name = "openstack-piyush-tf-1"

  cluster_profile {
    id = data.spectrocloud_cluster_profile.profile.id
  }

  cloud_account_id = data.spectrocloud_cloudaccount_openstack.account.id
  tags             = ["dev"]


  cloud_config {
    domain      = "Default"
    project     = "dev"
    region      = "RegionOne"
    ssh_key     = "Spectro2021"
    dns_servers = ["10.10.128.8", "8.8.8.8"]
    subnet_cidr = "192.168.151.0/24"
  }

  machine_pool {
    control_plane           = true
    control_plane_as_worker = true
    name                    = "master-pool"
    count                   = 1
    instance_type           = "spectro-xlarge"
    azs                     = ["zone1"]
  }

  machine_pool {
    name          = "worker-basic"
    count         = 2
    instance_type = "spectro-large"
    azs           = ["zone1"]
  }

}
