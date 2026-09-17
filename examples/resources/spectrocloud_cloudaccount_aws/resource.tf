# Nothing on this resource is ForceNew - every attribute below updates in place.
# Common to all three credential types below:
#   name                     - Required.
#   context                  - Optional, default "project". Allowed: "project", "tenant".
#   private_cloud_gateway_id - Optional. Set this when connecting through a Private Cloud
#                              Gateway to a private cluster endpoint.
#   partition                - Optional, default "aws". Set to "aws-us-gov" for GovCloud.
#   policy_arns              - Optional. Extra IAM policy ARNs to attach, beyond what Palette
#                              attaches automatically.

# Example 1: AWS Cloud Account with secret access/secret key credentials
#   type                   - Optional, default "secret". Allowed: "secret", "sts", "pod-identity".
#   aws_secured_access_key - Optional, sensitive. Preferred over the deprecated `aws_access_key`.
#                            The schema's own description calls the two mutually exclusive, but
#                            that check is currently commented out in the provider (toAwsAccount)
#                            - setting both won't error, the provider just prefers this field and
#                            silently ignores `aws_access_key`. Still best practice to set only
#                            one.
#   aws_secret_key         - Optional, sensitive. Used together with the access key above.
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

# Example 2: AWS Cloud Account with STS (role assumption)
#   arn         - Optional (Required in practice for type = "sts"). The role ARN to assume.
#   external_id - Optional, sensitive. Used alongside `arn` for cross-account role assumption.
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
#   role_arn                - Optional (Required in practice for type = "pod-identity"). The IAM
#                             role ARN used for EKS Pod Identity.
#   permission_boundary_arn - Optional. Caps the maximum permissions of roles Palette creates on
#                             your behalf.
resource "spectrocloud_cloudaccount_aws" "aws_pod_identity" {
  name                    = "aws-account-pod-identity"
  type                    = "pod-identity"
  role_arn                = var.aws_pod_identity_role_arn
  permission_boundary_arn = var.aws_permission_boundary_arn

  # Optional: Specify partition
  partition = "aws"

  # Optional: Additional policy ARNs
  policy_arns = ["arn:aws:iam::1234567890:policy/EKSPodIdentityPolicy"]
}
