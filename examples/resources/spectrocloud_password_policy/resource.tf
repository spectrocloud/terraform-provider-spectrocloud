resource "spectrocloud_password_policy" "policy_regex" {
  #  password_regex    = "*"
  password_expiry_days   = 999
  first_reminder_days    = 5
  min_password_length    = 6
  min_digits             = 1
  min_lowercase_letters  = 1
  min_special_characters = 1
  min_uppercase_letters  = 1
}

## import existing password policy
#import {
#  to = spectrocloud_password_policy.password_policy
#  id = "password-policy" // tenant-uid
#}