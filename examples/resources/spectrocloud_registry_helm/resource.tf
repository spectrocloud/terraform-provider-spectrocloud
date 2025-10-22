resource "spectrocloud_registry_helm" "r1" {
  name       = "us-artifactory"
  endpoint   = "https://123456.dkr.ecr.us-west-1.amazonaws.com"
  is_private = true
  credentials {
    credential_type = "noAuth"
    username        = "abc"
    password        = "def"
  }
  # Optional: Wait for the registry to complete synchronization
  # wait_for_sync = true
}