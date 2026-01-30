resource "spectrocloud_registry_oci" "r1" {
  name       = "test-nik2"
  type       = "ecr" # basic
  endpoint   = "123456.dkr.ecr.us-west-1.amazonaws.com"
  is_private = true
  credentials {
    credential_type = "sts"
    arn             = "arn:aws:iam::123456:role/stage-demo-ecr"
    external_id     = "sofiwhgowbrgiornM="
  }
}
# 
# Import by Name:
# terraform import spectrocloud_registry_oci.example "Pack Registry"
# terraform import spectrocloud_registry_oci.example "REGISTRY-NAME"

# Import by UID:
# terraform import spectrocloud_registry_oci.example "IMPORT-UID"
