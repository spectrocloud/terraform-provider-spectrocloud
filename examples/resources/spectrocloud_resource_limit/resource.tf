resource "spectrocloud_resource_limit" "resource_limit" {
  alert                  = 101
  api_keys               = 201
  appliance              = 6001
  appliance_token        = 201
  application_deployment = 200
  application_profile    = 200
  certificate            = 20
  cloud_account          = 355
  cluster                = 300
  cluster_group          = 50
  cluster_profile        = 2500
  filter                 = 200
  location               = 100
  macro                  = 6000
  private_gateway        = 100
  project                = 200
  registry               = 200
  role                   = 100
  ssh_key                = 300
  team                   = 100
  user                   = 300
  workspace              = 60
}

## import existing resource limit
#import {
#  to = spectrocloud_resource_limit.resource_limit
#  id = "5eea74e919f5e0d43fd3f316" // tenant-uid
#}