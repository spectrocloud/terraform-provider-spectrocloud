# Spectro Cloud credentials
sc_host         = "{Enter Spectro Cloud API Host}" # e.g: api.spectrocloud.com (for SaaS)
sc_api_key      = "{Enter Spectro Cloud API Key}"
sc_project_name = "{Enter Spectro Cloud Project Name}" # e.g: Default

# Cluster Configuration
cluster_name                 = "apache-cloudstack-cluster-1"
cluster_cloud_account_name   = "apache-cloudstack-account-1"
cluster_cluster_profile_name = "cloudstack-k8s-profile"

# CloudStack Configuration
cloudstack_zone_name               = "Zone1"
cloudstack_network_name            = "DefaultNetwork"
cloudstack_compute_offering        = "Medium Instance" # For control plane
cloudstack_compute_offering_worker = "Large Instance"  # For worker nodes

# Optional: SSH Key for node access
ssh_key_name = "my-ssh-key"

# Optional: Static IP Pool (if using static IPs)
# static_ip_pool_id = "<Enter Static IP Pool ID>"

# Optional: Backup Storage Location
# backup_storage_location_name = "s3-backup-location"

