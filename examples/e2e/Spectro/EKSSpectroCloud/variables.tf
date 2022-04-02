# Core Deployment Information
variable "env" { type = string }
variable "application" { type = string }
variable "uai" { type = string }
variable "aws_region" { type = string }
variable "aws_az_count" { type = number }
variable "aws_availability_zones" { type = list(any) }

# Virtual Network Infomration
variable "vpc_cidr" { type = string }
variable "trusted_cidrs" { type = list(any) }
variable "IPSubnets" { type = map(any) }

# Locals Brought Over
variable "ComponentName" { type = string }
variable "taggingstandard" { type = map(any) }
variable "aws_vpc_main_id" { type = string }
variable "aws_route_table_internetnat_id" { type = list(any) }

#SpectroCloud Variables
#variable SpectroCloudProject { type = string }
variable "SpectroCloudClusterProfiles" { type = map(any) }
variable "SpectroCloudAccount" { type = string }
variable "SpectroConfig" { type = map(any) }
variable "aws_subnets" { type = map(any) }