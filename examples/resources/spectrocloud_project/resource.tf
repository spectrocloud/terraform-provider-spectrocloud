# Day-2 mutability: nothing on this resource is ForceNew - name, tags, and description all
# update in place.
resource "spectrocloud_project" "project" {
  # Required. The project's name.
  name = "dev1"

  # Optional. Tags to assign to the project.
  tags = ["dev", "department:devops", "owner:bob"]

  # Optional. A human-readable description of the project.
  description = "Development project"
}