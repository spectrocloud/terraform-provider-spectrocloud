# Manages the tenant-wide password policy (singleton per tenant). Day-2 mutability: nothing here
# is ForceNew - every attribute updates in place.
#
# `password_regex` and the `min_*` requirements are mutually exclusive - setting password_regex
# to a non-empty value while any min_* field is non-zero fails validation. Use either a custom
# regex (uncomment below) or the individual minimum-requirement fields shown here, not both.
resource "spectrocloud_password_policy" "policy_regex" {
  # Optional, default "". Custom regex for password patterns - conflicts with the min_* fields.
  # password_regex = "^(?=.*[A-Z])(?=.*[a-z])(?=.*\\d).{8,}$"

  # Optional, default 999. Days before a password expires. Allowed range: 1-1000.
  password_expiry_days = 999
  # Optional, default 5. Days before expiry to send the first reminder to the user.
  first_reminder_days = 5
  # Optional, default 6. Minimum overall password length.
  min_password_length = 6
  # Optional. Minimum number of numeric digits (0-9) required.
  min_digits = 1
  # Optional. Minimum number of lowercase letters (a-z) required.
  min_lowercase_letters = 1
  # Optional. Minimum number of special characters (e.g. !@#$%) required.
  min_special_characters = 1
  # Optional. Minimum number of uppercase letters (A-Z) required.
  min_uppercase_letters = 1
}

## import existing password policy
#import {
#  to = spectrocloud_password_policy.password_policy
#  id = "password-policy" // tenant-uid
#}