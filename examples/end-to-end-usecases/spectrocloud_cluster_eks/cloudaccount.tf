# Creates and manages a dedicated AWS cloud account for this end-to-end example, rather than
# looking up one that already exists - so this folder is fully self-contained.
#
# Day-2 mutability: nothing on this resource is ForceNew - every attribute updates in place.
resource "spectrocloud_cloudaccount_aws" "account" {
  name    = "e2e-eks-account"
  context = "project"

  type                   = "secret"
  aws_secured_access_key = var.aws_access_key
  aws_secret_key         = var.aws_secret_key

  # STS alternative (mutually exclusive with the secret-key fields above; set type = "sts"):
  # arn         = var.aws_sts_role_arn
  # external_id = var.aws_sts_external_id

  # EKS Pod Identity alternative (set type = "pod-identity") - particularly relevant for EKS,
  # since it lets pods assume IAM roles without node-level credentials:
  # role_arn                = var.aws_pod_identity_role_arn
  # permission_boundary_arn = var.aws_permission_boundary_arn

  partition   = "aws"
  policy_arns = []
}

# Alternative to creating a new account: look up one that's already registered in Palette.
# data "spectrocloud_cloudaccount_aws" "account" {
#   name = "some-existing-account-name"
# }
