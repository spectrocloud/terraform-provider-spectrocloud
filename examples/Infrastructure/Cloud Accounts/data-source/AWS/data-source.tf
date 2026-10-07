# Looks up an existing AWS cloud account registered in Palette, by name or by ID.

# Retrieve details of an AWS cloud account using name.
#
# Lookup keys:
#   name    - Exactly one of `id`/`name` required, also Computed.
#   context - Optional. Allowed: "project", "tenant", "" (default). Required only to
#             disambiguate when more than one account shares the same `name` across scopes.
data "spectrocloud_cloudaccount_aws" "example" {
  name    = "example-aws-account"
  context = "project"
}

# Retrieve details of an AWS cloud account using ID.
#
# Lookup keys:
#   id - Exactly one of `id`/`name` required, also Computed.
data "spectrocloud_cloudaccount_aws" "by_id" {
  id = "123e4567-e89b-12d3-a456-426614174000"
}

# Output cloud account details
output "aws_account_id" {
  value = data.spectrocloud_cloudaccount_aws.example.id
}

output "aws_account_name" {
  value = data.spectrocloud_cloudaccount_aws.example.name
}

output "aws_account_context" {
  value = data.spectrocloud_cloudaccount_aws.example.context
}
