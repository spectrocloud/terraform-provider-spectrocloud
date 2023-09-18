
resource "spectrocloud_cluster_aws" "cluster" {

  name = "lmco-test"

  # Tags to be set in AWS for the cluster
  tags = ["owner:bob"]

  # Infra profile
  cluster_profile {
    id = data.spectrocloud_cluster_profile.infra_demo.id
  }

  # Addon profile
  cluster_profile {
    id = spectrocloud_cluster_profile.addon_profile.id
  }

  # Cloud account to use for cluster provisioning
  cloud_account_id = data.spectrocloud_cloudaccount_aws.demo_cloudaccount.id

  cloud_config {
    ssh_key_name     = var.aws_ssh_key_name
    control_plane_lb = var.control_plane_lb
    region           = "us-gov-east-1"
    vpc_id           = "vpc-04906e2f9614976bf"
  }

  scan_policy {
    configuration_scan_schedule = "0 0 * * SUN"
    penetration_scan_schedule   = "0 0 * * SUN"
    conformance_scan_schedule   = "0 0 1 * *"
  }

  machine_pool {
    # Labels to be set on the nodes in this pool
    additional_labels = {
      "department" = "eng"
    }
    control_plane = true
    #control_plane_as_worker = true
    #additional_security_groups = []
    name          = "master-pool"
    count         = 1
    instance_type = "t3.xlarge"
    disk_size_gb  = 60
    az_subnets = {
      "us-gov-east-1a" = "subnet-017ef5ac4267b5730,subnet-053f4a19e6b358731"
    }
  }

  machine_pool {
    # Labels to be set on the nodes in this pool
    additional_labels = {
      "type" = "worker"
    }
    #additional_security_groups = []
    name          = "worker-pool"
    count         = 1
    instance_type = "t3.xlarge"
    disk_size_gb  = 60
    az_subnets = {
      "us-gov-east-1a" = "subnet-017ef5ac4267b5730"
    }
  }

}
