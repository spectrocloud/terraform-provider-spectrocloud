variable "cluster_profile_name" {
  type        = string
  description = "The name to give the cluster profile that will be created."
  default     = "pack-tutorial-profile"
}

variable "cluster_profile_description" {
  type        = string
  description = "A human-readable description for the cluster profile."
  default     = "My cluster profile as part of the packs tutorial."
}

variable "cluster_name" {
  type        = string
  description = "The name to give the AWS cluster that will be created."
  default     = "pack-tutorial-cluster"
}

variable "instance_type" {
  type        = string
  description = "The AWS EC2 instance type to use for both the control plane and worker nodes."
  default     = "m4.xlarge"
}

# ToDo: Provide a value for the variable below. The value will be the actual cloud account name added to your Palette project settings.
variable "cluster_cloud_account_aws_name" {
  type        = string
  description = "The name of the AWS cloud account already registered in your Palette project settings (Tenant Settings > Cloud Accounts)."
}

# ToDo: Provide a value for the variable below. The value will be one of the [AWS regions](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html)
# The tutorial example uses "us-east-1" region.
variable "aws_region_name" {
  type        = string
  description = "The AWS region to deploy the cluster into, for example \"us-east-1\"."
}

# ToDo: Provide a value for the variable below. The value will be one of the [AWS Availability Zones](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html)
# The tutorial example uses "us-east-1a" availability zone.
variable "aws_az_names" {
  type        = list(string)
  description = "The AWS Availability Zones to deploy into, for example [\"us-east-1a\", \"us-east-1b\"]. If left empty, the first available AZ in aws_region_name is used automatically."
  default     = []
}

# ToDo: Provide a value for the variable below. The value will be the SSH key created in the AWS region where you will deploy the cluster.
variable "ssh_key_name" {
  type        = string
  description = "The name of an existing AWS EC2 key pair in aws_region_name, used for SSH access to the cluster nodes."
}

# ToDo: Provide the name of your private registry server.
# The tutorial example uses "private-pack-registry".
variable "private_pack_registry" {
  type        = string
  description = "The name, in Palette, of the pack registry that hosts the custom add-on pack from the tutorial (used only when use_oci_registry is false)."
}

variable "custom_addon_pack" {
  type        = string
  description = "The name of the custom add-on pack (created in the tutorial) to add to the cluster profile."
  default     = "hellouniverse"
}

variable "custom_addon_pack_version" {
  type        = string
  description = "The version of the custom add-on pack to add to the cluster profile."
  default     = "1.0.0"
}

# ToDo: Set the use of OCI registry to true or false.
# The default value is set as true.
variable "use_oci_registry" {
  type        = bool
  description = "Whether the custom add-on pack is hosted in an OCI registry (true) or a standard Palette pack registry (false)."
  default     = true
}

variable "tags" {
  type        = list(string)
  description = "Tags applied to the Palette resources created by this example. Each tag must be 63 characters or fewer, start and end with an alphanumeric character, and contain only alphanumeric characters, dots, dashes, or underscores (no slashes)."
  default     = ["spectro-cloud-education", "app:hello-universe", "terraform_managed:true"]
}

locals {
  # If aws_az_names is left empty, fall back to the first AZ Terraform discovers as available in aws_region_name.
  azs = length(var.aws_az_names) != 0 ? var.aws_az_names : slice(data.aws_availability_zones.available.names, 0, 1)
}
