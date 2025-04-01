# Retrieve details of an AWS cloud account using name
data "spectrocloud_cloud_account_aws" "example" {
  name    = "example-aws-account"  # Required if 'id' is not provided
  context = "project"              # Optional: Allowed values are "project", "tenant", or "" (default)
}

# Retrieve details of an AWS cloud account using ID
data "spectrocloud_cloud_account_aws" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000"  # Required if 'name' is not provided
}

# Output cloud account details
output "aws_account_id" {
  value = data.spectrocloud_cloud_account_aws.example.id
}

output "aws_account_name" {
  value = data.spectrocloud_cloud_account_aws.example.name
}

output "aws_account_context" {
  value = data.spectrocloud_cloud_account_aws.example.context
}