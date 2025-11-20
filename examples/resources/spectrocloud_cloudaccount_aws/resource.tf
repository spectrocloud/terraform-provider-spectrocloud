# Example 1: AWS Cloud Account with Secret Credentials
resource "spectrocloud_cloudaccount_aws" "aws_secret" {
  name                   = "aws-account-secret"
  type                   = "secret"
  aws_secured_access_key = var.aws_secured_access_key # or aws_access_key=<access_key>
  aws_secret_key         = var.aws_secret_key

  # If US GOV partition needs to be used uncomment the below line
  # partition = "aws-us-gov"

  # Additional policies can be added to the below list
  policy_arns = ["arn:aws:iam::1234567890:policy/AWSLoadBalancerControllerIAMPolicy"]
}

# Example 2: AWS Cloud Account with STS (Role Assumption)
resource "spectrocloud_cloudaccount_aws" "aws_sts" {
  name        = "aws-account-sts"
  type        = "sts"
  arn         = var.aws_sts_role_arn
  external_id = var.aws_external_id

  # Optional: Specify partition
  partition = "aws"

  # Optional: Additional policy ARNs
  policy_arns = ["arn:aws:iam::1234567890:policy/CustomSTSPolicy"]
}

# Example 3: AWS Cloud Account with EKS Pod Identity
resource "spectrocloud_cloudaccount_aws" "aws_pod_identity" {
  name                    = "aws-account-pod-identity"
  type                    = "pod-identity"
  role_arn                = var.aws_pod_identity_role_arn
  permission_boundary_arn = var.aws_permission_boundary_arn # Optional

  # Optional: Specify partition
  partition = "aws"

  # Optional: Additional policy ARNs
  policy_arns = ["arn:aws:iam::1234567890:policy/EKSPodIdentityPolicy"]
}
