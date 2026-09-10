# Manages tenant-wide Palette object quota limits (singleton per tenant). Day-2 mutability:
# nothing here is ForceNew - every limit updates in place. All attributes are optional; each
# defaults to the value Palette applies out of the box (shown below) if omitted. Every limit
# accepts 1-10,000 except `appliance` and `cluster`, which allow up to 50,000.
resource "spectrocloud_resource_limit" "resource_limit" {
  alert                  = 101  # default 100
  api_keys               = 201  # default 20
  appliance              = 6001 # default 200, range 1-50000
  appliance_token        = 201  # default 200
  application_deployment = 200  # default 100
  application_profile    = 200  # default 100
  certificate            = 20   # default 20
  cloud_account          = 355  # default 200
  cluster                = 300  # default 10000, range 1-50000
  cluster_group          = 50   # default 100
  cluster_profile        = 2500 # default 200
  filter                 = 200  # default 100
  location               = 100  # default 100
  macro                  = 6000 # default 200
  private_gateway        = 100  # default 50
  project                = 200  # default 50
  registry               = 200  # default 50
  role                   = 100  # default 100
  ssh_key                = 300  # default 300
  team                   = 100  # default 100
  user                   = 300  # default 300
  workspace              = 60   # default 50
}

## import existing resource limit
#import {
#  to = spectrocloud_resource_limit.resource_limit
#  id = "5eea74e919f5e0d43fd3f316" // tenant-uid
#}