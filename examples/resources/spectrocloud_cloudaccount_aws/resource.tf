resource "spectrocloud_cloudaccount_aws" "aws-1" {
  name = "aws-1"
  #If US GOV partition needs to be used uncomment the below line
  #partition      = "aws-us-gov"

  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key
  type           = "secret"

  #Additional policies can be added to the below list
  policy_arns = ["arn:aws:iam::1234567890:policy/AWSLoadBalancerControllerIAMPolicy"]
}
