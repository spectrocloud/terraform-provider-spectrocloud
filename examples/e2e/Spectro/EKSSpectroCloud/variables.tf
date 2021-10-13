# Core Deployment Information
variable env { type = string }
variable application { type = string }
variable uai { type = string }
variable aws_region { type = string }
variable aws_az_count { type = number }
variable aws_availability_zones { type = list }

# Virtual Network Infomration
variable vpc_cidr { type = string }
variable trusted_cidrs { type = list }
variable IPSubnets { type = map }

# Locals Brought Over
variable ComponentName { type = string }
variable taggingstandard { type = map }
variable aws_vpc_main_id { type = string }
variable aws_route_table_internetnat_id { type = list }

#SpectroCloud Variables
#variable SpectroCloudProject { type = string }
variable SpectroCloudClusterProfiles { type = map }
variable SpectroCloudAccount { type = string }
variable SpectroConfig { type = map }
variable aws_subnets { type = map }