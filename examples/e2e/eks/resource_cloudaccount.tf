#
# If looking up a cloudaccount instead of creating one
//data "spectrocloud_cloudaccount_aws" "account" {
//  name = "aws-dev"
//}

resource "spectrocloud_cloudaccount_aws" "account" {
  name            = "aws-dev-access"
  aws_access_key  = var.aws_access_key
  aws_secret_key  = var.aws_secret_key
  type            = var.cloud_account_type
  arn             = var.arn
  external_id     = var.external_id
}
