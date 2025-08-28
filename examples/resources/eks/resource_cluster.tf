resource "spectrocloud_cluster_eks" "cluster" {
  name = "ran-tf-eks"
  context = "tenant"
  tags_map = {"QA" = "ranjithroy@_ .:/=+-@.@123"}

  timeouts {
    create = "5m"
    update = "30m"
    delete = "59m"
  }

  cluster_profile {
    id = "68a6e0bc500766a5c9241784"
  }

#  cluster_profile {
#    id = "67d07f687d0bcbd1fe2bcc72"
#  }



#  cluster_profile {
#    id = spectrocloud_cluster_profile.profile-rbac.id
#    pack {
#      name   = "spectro-rbac"
#      tag    = "1.0.0"
#      values = file("rbac.yaml")
#    }
#  }

  cloud_account_id = "68a6e0ec788fd02b1e0151a4"

#   cluster_rbac_binding {
#    type      = "RoleBinding"
#    namespace = "test5ns"
#    role = {
#      kind = "Role"
#      name = "testrolefromns3"
#    }
#    subjects {
#      type = "User"
#      name = "ranjith.p@spectrocloud.com"
#    }
#    subjects {
#      type = "Group"
#      name = "testGroupFromNS3"
#    }
#    subjects {
#      type      = "ServiceAccount"
#      name      = "testrolesubject3"
#      namespace = "testrolenamespace"
#    }
#  }

  cloud_config {
    ssh_key_name = var.aws_ssh_key_name
    region       = var.aws_region
    vpc_id       = var.aws_vpc_id
    azs          = ["ap-south-1a", "ap-south-1b"]
#    az_subnets   = var.cp_azs_subnets_map != {} ? var.cp_azs_subnets_map : null
  }

  machine_pool {
    name          = "worker-basic"
    count         = 1
    instance_type = "t3.xlarge"
    azs          = ["ap-south-1a", "ap-south-1b", "ap-south-1c"]
#    az_subnets    = var.cp_azs_subnets_map != {} ? var.cp_azs_subnets_map : null
    disk_size_gb  = 60
  }

  machine_pool {
    name          = "worker-basic-2"
    count         = 1
    instance_type = "t3.xlarge"
    azs          = ["ap-south-1a", "ap-south-1b", "ap-south-1c"]
#    az_subnets    = var.cp_azs_subnets_map != {} ? var.cp_azs_subnets_map : null
    disk_size_gb  = 60
  }
#
#  fargate_profile {
#    name    = "fg-1"
#    subnets = values(var.worker_azs_subnets_map)
#    additional_tags = {
#      hello = "yo"
#    }
#    selector {
#      namespace = "fargate"
#      labels = {
#        abc = "cool"
#      }
#
#    }
#  }

}



#data "spectrocloud_cloudaccount_aws" "account" {
#  # id = <uid>
#  name = "ran-tf-2"
#  context = "project"
#}
#
#data "spectrocloud_cluster_profile" "profile" {
#  # id = <uid>
#  name = "tf-eks-profile"
#}
#
#resource "spectrocloud_cluster_eks" "cluster" {
#  name             = "ran-eks-tf-2"
#  tags             = ["dev", "department:qa", "owner:roy"]
#  cloud_account_id = data.spectrocloud_cloudaccount_aws.account.id
#
#  cloud_config {
#    ssh_key_name = var.aws_ssh_key_name
#    region       = "us-east-1"
#    azs          = ["us-east-1a","us-east-1f"]
#  }
#
#  cluster_profile {
#    id = data.spectrocloud_cluster_profile.profile.id
#  }
#
##  backup_policy {
##    schedule                  = "0 0 * * SUN"
##    backup_location_id        = data.spectrocloud_backup_storage_location.bsl.id
##    prefix                    = "prod-backup"
##    expiry_in_hour            = 7200
##    include_disks             = true
##    include_cluster_resources = true
##  }
#
##  scan_policy {
##    configuration_scan_schedule = "0 0 * * SUN"
##    penetration_scan_schedule   = "0 0 * * SUN"
##    conformance_scan_schedule   = "0 0 1 * *"
##  }
#
#  machine_pool {
#    name          = "worker-basic"
#    count         = 1
#    instance_type = "t3.xlarge"
#    disk_size_gb  = 60
#    azs          = ["us-east-1b","us-east-1c"]
#  }
#
#}