# Looks up a backup storage location by name or by ID.
#
# Lookup keys (exactly one of `id`/`name` required, both also Computed):
#   name - Name of the backup storage location.
#   id   - ID of the backup storage location.
data "spectrocloud_backup_storage_location" "example" {
  name = "my-backup-location"
  # id = "657ec9a27afca71b0dc98027"
}

# Computed.
output "backup_storage_location_id" {
  value = data.spectrocloud_backup_storage_location.example.id
}

output "backup_storage_location_name" {
  value = data.spectrocloud_backup_storage_location.example.name
}
