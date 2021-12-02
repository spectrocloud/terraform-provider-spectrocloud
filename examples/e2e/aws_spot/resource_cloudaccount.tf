#
# If looking up a cloudaccount instead of creating one
# data "spectrocloud_cloudaccount_aws" "account" {
#   # id = <uid>
#   name = var.cluster_cloud_account_name
# }

resource "spectrocloud_cloudaccount_aws" "account" {
  name           = "aws-picard-4"
  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
}
