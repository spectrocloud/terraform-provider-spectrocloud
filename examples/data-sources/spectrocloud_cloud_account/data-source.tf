data "spectrocloud_cloudaccount_aws" "aws_account" {
  # id = <uid>
  name = "srini-aws-sts"
  context = "tenant"
}

output "aws_account" {
  value = data.spectrocloud_cloudaccount_aws.aws_account
}