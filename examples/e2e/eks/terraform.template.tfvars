# Spectro Cloud credentials
sc_host         = "{enter Spectro Cloud API endpoint}" #e.g: api.spectrocloud.com (for SaaS)
sc_username     = "{enter Spectro Cloud username}"     #e.g: user1@abc.com
sc_password     = "{enter Spectro Cloud password}"     #e.g: supereSecure1!
sc_project_name = "{enter Spectro Cloud project Name}" #e.g: Default


# AWS Cloud Account credentials
# Ensure minimum AWS account permissions:
# https://docs.spectrocloud.com/clusters/?clusterType=aws_cluster#awscloudaccountpermissions

# Specify Access Key & Secret Key OR specify STS information
cloud_account_type = "{enter AWS Cloud Account Type}" #eg. "secret" or "sts"

# OPTION 1 : Specifying access key and secret key where cloud_account_type="secret"
aws_access_key = "{enter AWS access key}"
aws_secret_key = "{enter AWS secret key}"

# OPTION 2 : Specifying arn and external id where cloud_account_type="sts"
arn         = "{enter AWS Arn}"
external_id = "{enter AWS External Id}"


# Existing SSH Key in AWS
# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
aws_ssh_key_name = "{enter AWS SSH key name}" #e.g: default

# Enter the AWS Region and AZ for cluster resources
# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions
aws_region = "{enter AWS Region}" #e.g: us-west-2

aws_vpc_id = "{enter AWS Vpc Id}" #e.g: vpc-0e03ff84a894d4

master_azs_subnets_map = {
  "{enter AWS Availability Zone A}" = "{enter Subnet For Availability Zone A}",
  "{enter AWS Availability Zone B}" = "{enter Subnet For Availability Zone B}"
}
/*
eg. master_azs_subnets_map = {
      "us-west-2a" = "subnet-0d4978ddbff16c868",
      "us-west-2b" = "subnet-041a35c9c06eeb701"
    }
*/

worker_azs_subnets_map = {
  "{enter AWS Availability Zone A}" = "{enter Subnet For Availability Zone A}",
  "{enter AWS Availability Zone B}" = "{enter Subnet For Availability Zone B}"
}