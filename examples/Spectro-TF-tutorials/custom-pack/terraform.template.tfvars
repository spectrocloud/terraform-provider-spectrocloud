cluster_cloud_account_aws_name = "REPLACE ME" # Name of the AWS cloud account already registered in your Palette project settings
aws_region_name                = "REPLACE ME" # An AWS region, e.g. "us-east-1"
aws_az_names                   = []           # Optional: AWS Availability Zones to use, e.g. ["us-east-1a", "us-east-1b"]. Leave empty to auto-select one.
ssh_key_name                   = "REPLACE ME" # Name of an existing AWS EC2 key pair in aws_region_name
private_pack_registry          = "REPLACE ME" # Name, in Palette, of the registry hosting the custom add-on pack from the tutorial
use_oci_registry               = true         # true if private_pack_registry is an OCI registry, false for a standard Palette pack registry
