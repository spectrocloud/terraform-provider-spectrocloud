resource "spectrocloud_cloudaccount_aws" "aws-1" {
  name                = "aws-1"
  aws_access_key     = var.aws_access_key
  aws_secret_key     = var.aws_secret_key
}
