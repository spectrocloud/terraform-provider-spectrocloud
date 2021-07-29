/*
resource "spectrocloud_cloudaccount_openstack" "account" {
  name           = "openstack-dev"
  private_cloud_gateway_id = "60fe915794168c655c0d766a"
  openstack_username = var.openstack_username
  openstack_password = var.openstack_password
  identity_endpoint = var.identity_endpoint
  parent_region = var.region
  default_domain = var.domain
  default_project = var.project
}
*/

data "spectrocloud_cloudaccount_openstack" "account" {
  name = "openstack-pcg-piyush-dev-2"
}
