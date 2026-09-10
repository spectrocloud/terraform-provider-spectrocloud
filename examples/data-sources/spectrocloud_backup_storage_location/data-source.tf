# Looks up a backup storage location by name or by ID.
data "spectrocloud_backup_storage_location" "example" {
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  name = "my-backup-location"
  # Lookup key, optional (exactly one of `id`/`name` required), also Computed.
  # id = "657ec9a27afca71b0dc98027"
}

# Computed.
output "backup_storage_location_id" {
  value = data.spectrocloud_backup_storage_location.example.id
}

output "backup_storage_location_name" {
  value = data.spectrocloud_backup_storage_location.example.name
}
