resource "spectrocloud_registration_token" "tf_token" {
  name        = "tf_siva"
  description = "test token description updated"
  expiry_date = "2025-03-25"
  project_uid = "6514216503b"
  status      = "active"
}

## import existing registration token
#import {
#  to = spectrocloud_registration_token.token
#  id = "{tokenUID}" //tokenUID
#}
