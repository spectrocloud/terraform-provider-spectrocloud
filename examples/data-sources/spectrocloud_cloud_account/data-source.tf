data "spectrocloud_cloudaccount_aws" "aws_account" {
  # id = <uid>
  name = "srini-aws-sts"
  context = "tenant"
}

output "same" {
  value = data.spectrocloud_cloudaccount_aws.aws_account
}