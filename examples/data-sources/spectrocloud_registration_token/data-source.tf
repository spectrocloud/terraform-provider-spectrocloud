data "spectrocloud_registration_token" "tf" {
  name = "ran-dev-test"
  #  id = "657ec9a27afca71b0dc98027"
}

output "token" {
  value = data.spectrocloud_registration_token.tf.token
}
