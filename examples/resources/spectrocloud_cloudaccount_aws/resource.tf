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
resource "spectrocloud_cloudaccount_aws" "aws_secret" {
  name = "aws-account-secret"
  # Optional, default "secret". Allowed: "secret", "sts", "pod-identity".
  type = "secret"
  # Optional, sensitive. Preferred over the deprecated `aws_access_key`, which is mutually
  # exclusive with this field.
  aws_secured_access_key = var.aws_secured_access_key # or aws_access_key=<access_key>
  # Optional, sensitive. Used together with the access key above.
  aws_secret_key = var.aws_secret_key

  # If US GOV partition needs to be used uncomment the below line
  # partition = "aws-us-gov"

  # Additional policies can be added to the below list
  policy_arns = ["arn:aws:iam::1234567890:policy/AWSLoadBalancerControllerIAMPolicy"]
}

# Example 2: AWS Cloud Account with STS (role assumption)
resource "spectrocloud_cloudaccount_aws" "aws_sts" {
  name = "aws-account-sts"
  type = "sts"
  # Optional (Required in practice for type = "sts"). The role ARN to assume.
  arn = var.aws_sts_role_arn
  # Optional, sensitive. Used alongside `arn` for cross-account role assumption.
  external_id = var.aws_external_id

  # Optional: Specify partition
  partition = "aws"

  # Optional: Additional policy ARNs
  policy_arns = ["arn:aws:iam::1234567890:policy/CustomSTSPolicy"]
}

# Example 3: AWS Cloud Account with EKS Pod Identity
resource "spectrocloud_cloudaccount_aws" "aws_pod_identity" {
  name = "aws-account-pod-identity"
  type = "pod-identity"
  # Optional (Required in practice for type = "pod-identity"). The IAM role ARN used for EKS
  # Pod Identity.
  role_arn = var.aws_pod_identity_role_arn
  # Optional. Caps the maximum permissions of roles Palette creates on your behalf.
  permission_boundary_arn = var.aws_permission_boundary_arn # Optional

  # Optional: Specify partition
  partition = "aws"

  # Optional: Additional policy ARNs
  policy_arns = ["arn:aws:iam::1234567890:policy/EKSPodIdentityPolicy"]
}
