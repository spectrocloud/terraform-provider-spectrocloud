
#SpectroCloud Variables
variable "SpectroCloudProject" { type = string }
variable "SpectroCloudAccount" { type = string }
variable "SpectroCloudUsername" { type = string }
variable "SpectroCloudPassword" { type = string }
variable "SpectroCloudURI" { type = string }
variable "SpectroConfig" { type = map(any) }
variable "SpectroCloudClusterProfiles" { type = map(any) }

module "EKSSpectroCloud" {
  source        = "../EKSSpectroCloud"
  ComponentName = "EKSCluster"

  # Core Deployment Information
  env                    = var.env
  application            = var.application
  uai                    = var.uai
  aws_region             = var.aws_region
  aws_az_count           = var.aws_az_count
  aws_availability_zones = var.aws_availability_zones

  # Virtual Network Infomration
  vpc_cidr      = var.vpc_cidr
  IPSubnets     = local.IPSubnets
  trusted_cidrs = var.trusted_cidrs

  # Locals Brought Over
  taggingstandard = local.taggingstandard

  #Dependancy map to prevent NAT Gateway and Internet Gateway from being pulled from EKS preventing cluster destruction
  depends_on = [
    module.core.aws_route_table_internetnat_id, module.core.aws_route_table_internetigw_id, module.core.aws_nat_gateway_ngw_id, module.EKSFramework.EKS_Subnets, module.core.aws_route_internetnat, module.EKSFramework.aws_route_table_association_k8snatgw, module.core.aws_internet_gateway_igw_id, module.core.aws_route_internetigw_id, module.core.aws_route_table_association_publicinternet_id
  ]

  #Module Outputs
  aws_vpc_main_id                = module.core.aws_vpc_main_id
  aws_route_table_internetnat_id = module.core.aws_route_table_internetnat_id

  #SpectroCloud Variables
  #SpectroCloudProject = var.SpectroCloudProject
  SpectroCloudClusterProfiles = var.SpectroCloudClusterProfiles
  SpectroCloudAccount         = var.SpectroCloudAccount
  SpectroConfig               = var.SpectroConfig
  aws_subnets                 = module.EKSFramework.EKS_Subnets
}