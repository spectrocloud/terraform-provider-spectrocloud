data "spectrocloud_cloudaccount_aws" "aws_account" {
  # id = <uid>
  name    = "srini-aws-sts"
#   context = "project"
#   context = "tenant"
}

output "same" {
  value = data.spectrocloud_cloudaccount_aws.aws_account
}