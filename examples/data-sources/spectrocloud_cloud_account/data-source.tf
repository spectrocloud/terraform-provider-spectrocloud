data "spectrocloud_cloudaccount_aws" "aws_account" {
  # id = <uid>
  name    = "ran-tf"
  context = "tenant"
}

output "aws_account" {
  value = data.spectrocloud_cloudaccount_aws.aws_account
}