# Day-2 mutability: nothing on this resource is ForceNew - name, description, project_uid,
# expiry_date, and status all update in place.
resource "spectrocloud_registration_token" "tf_token" {
  # Required. The token's name.
  name = "tf_siva"
  # Optional, default "". A brief description of the token.
  description = "Registration token for edge host enrollment"
  # Required. Expiration date in YYYY-MM-DD format. Computed dynamically here (1 year out) so
  # this example never ships with a stale, already-expired date.
  expiry_date = formatdate("YYYY-MM-DD", timeadd(timestamp(), "8760h"))
  # Optional, default "". UID of the project to scope this token to. Leave unset to scope the
  # token at the tenant level.
  project_uid = "6514216503b"
  # Optional, default "active". Allowed: "active", "inactive".
  status = "active"
}

## import existing registration token
#import {
#  to = spectrocloud_registration_token.token
#  id = "{tokenUID}" //tokenUID
#}
