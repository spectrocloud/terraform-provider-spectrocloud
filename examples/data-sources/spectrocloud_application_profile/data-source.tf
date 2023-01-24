data "spectrocloud_application_profile" "app_postgres" {
  name    = "profile-postgres"
  version = "1.0.0"
}

data "spectrocloud_application_profile" "app_ingress" {
  name = "profile-gamebox-ingress"
}

output "out_ingress_version" {
  value = data.spectrocloud_application_profile.app_ingress.version
}

output "out_postgres" {
  value = data.spectrocloud_application_profile.app_postgres.id
}
