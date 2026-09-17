# Day-2 mutability: nothing on this resource is ForceNew - name, tags, and description all
# update in place.
#
# Attributes:
#   name        - Required. The project's name.
#   tags        - Optional. Tags to assign to the project.
#   description - Optional. A human-readable description of the project.
resource "spectrocloud_project" "project" {
  name        = "dev1"
  tags        = ["dev", "department:devops", "owner:bob"]
  description = "Development project"
}