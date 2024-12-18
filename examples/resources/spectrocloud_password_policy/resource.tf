resource "spectrocloud_password_policy" "policy_regex" {
  #  password_regex    = "*"
  password_expiry_days   = 123
  first_reminder_days    = 5
  min_digits             = 1
  min_lowercase_letters  = 12
  min_password_length    = 12
  min_special_characters = 1
  min_uppercase_letters  = 1
}