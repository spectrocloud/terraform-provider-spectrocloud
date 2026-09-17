# Manages tenant-wide Palette object quota limits (singleton per tenant). Day-2 mutability:
# nothing here is ForceNew - every limit updates in place. All attributes are optional; each
# defaults to the value Palette applies out of the box (shown below) if omitted. Every limit
# accepts 1-10,000 except `appliance` and `cluster`, which allow up to 50,000.
#
# Defaults (all fields optional):
#   alert                  - default 100
#   api_keys               - default 20
#   appliance              - default 200, range 1-50000
#   appliance_token        - default 200
#   application_deployment - default 100
#   application_profile    - default 100
#   certificate            - default 20
#   cloud_account          - default 200
#   cluster                - default 10000, range 1-50000
#   cluster_group          - default 100
#   cluster_profile        - default 200
#   filter                 - default 100
#   location               - default 100
#   macro                  - default 200
#   private_gateway        - default 50
#   project                - default 50
#   registry               - default 50
#   role                   - default 100
#   ssh_key                - default 300
#   team                   - default 100
#   user                   - default 300
#   workspace              - default 50
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