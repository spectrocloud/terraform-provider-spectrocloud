##################################################################################
# Spectro Cloud credentials
##################################################################################
sc_host         = "{Enter Spectro Cloud API Host}" #e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" #e.g: Default


##################################################################################
# Cluster Properties
##################################################################################

# Existing SSH Key in AWS (optional for STATIC provisioning)
# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
# For DYNAMIC provisioning, SSH key is required
# aws_ssh_key_name = "{enter AWS SSH key name}" #e.g: default

# Enter the AWS Region and AZ for cluster resources
# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions
aws_region = "{enter AWS Region}" #e.g: us-west-2

##################################################################################
# Cloud Account (Secret or STS)
##################################################################################
# AWS cloud account authentication supports either direct access_key/secret_key
# or via STS, which assumes a role in the user's cloud account.
# Instructions and AWS policies are available in the in-product help:
# 1. Navigate to Project -> Cloud Accounts
# 2. Click to add a new "AWS Cloud Account"
# 3. Toggle the appropriate authentication type: Secret or STS
# 4. Review the right-hand panel for instructions and information.
# Additional instructions available at:
# https://docs.spectrocloud.com/clusters/?clusterType=aws_cluster#awscloudaccountpermissions

# Specify cloud_account_type and uncomment option SECRET or option STS below

cloud_account_type = "{enter AWS Cloud Account Type}" #eg. "secret" or "sts"

#######################
# Option SECRET
#######################
# (for SECRET, uncomment the following 2 lines)
# aws_access_key = "{enter AWS access key}"
# aws_secret_key = "{enter AWS secret key}"

#######################
# Option STS
#######################
# (for STS, uncomment the following 2 lines)
# arn         = "{enter AWS Arn}"
# external_id = "{enter AWS External Id}"


########################
# Static Provisioning
########################
# Static provisioning requires specifying the exiting VPC-ID and all subnets to target
aws_vpc_id = "{enter AWS VPC ID}" #e.g: vpc-123456

master_azs_subnets_map = {
  "{enter AWS Availability Zone A}" = "{enter Subnet for AZ A, ...}"
  "{enter AWS Availability Zone B}" = "{enter Subnet for AZ A, ...}"
}
worker_azs_subnets_map = {
  "{enter AWS Availability Zone A}" = "{enter Subnet for AZ A, ...}"
  "{enter AWS Availability Zone B}" = "{enter Subnet for AZ A, ...}"
}
