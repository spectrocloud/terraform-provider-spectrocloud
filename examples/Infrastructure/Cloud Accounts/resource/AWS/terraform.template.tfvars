# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# Secret Credentials (Example 1)
aws_secured_access_key = "<Enter AWS Access Key>"
aws_secret_key         = "<Enter AWS Secret Key>"

# STS Credentials (Example 2)
aws_sts_role_arn = "arn:aws:iam::123456789012:role/SpectroCloudRole"
aws_external_id  = "<Enter External ID>"

# Pod Identity Credentials (Example 3)
aws_pod_identity_role_arn   = "arn:aws:iam::123456789012:role/EKSPodIdentityRole"
aws_permission_boundary_arn = "arn:aws:iam::123456789012:policy/PermissionBoundary"