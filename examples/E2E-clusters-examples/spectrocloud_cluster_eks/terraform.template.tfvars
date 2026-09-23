# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default

# AWS cloud account (created by this example, not looked up)
aws_access_key = "{enter AWS access key}"
aws_secret_key = "{enter AWS secret key}"

# Cluster placement
cluster_ssh_key_name   = "{enter existing EC2 key pair name}"
aws_region             = "us-east-1"
eks_fargate_subnet_ids = ["{enter subnet ID}", "{enter another subnet ID}"]

# Backup storage location (S3), created by this example
backup_s3_bucket_name = "{enter S3 bucket name}"
backup_s3_access_key  = "{enter AWS access key}"
backup_s3_secret_key  = "{enter AWS secret key}"

# Where to write the local kubeconfig files (defaults to the current directory)
# kubeconfig_output_dir = "."
