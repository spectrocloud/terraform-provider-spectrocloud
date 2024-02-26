resource "spectrocloud_registry_helm" "r1" {
  name       = "test-artifactory"
  endpoint   = "https://123456.dkr.ecr.us-west-1.amazonaws.com"
  is_private = true
#  mode = "app"
  credentials {
    credential_type = "basic"
    username        = "abc"
    password        = "def"
  }
}