# Cluster Profile Variables
metallb_ip = "REPLACE ME" # Provide a range of IP addresses for your Metallb Load Balancer. This range must be included in the PCG's static IP pool range if using static IP placement.

# Cluster Variables
pcg_name           = "REPLACE ME" # Provide the name of the PCG that will be used to deploy the Palette cluster.
datacenter_name    = "REPLACE ME" # Provide the name of the datacenter in vSphere.
folder_name        = "REPLACE ME" # Provide the name of the folder in vSphere.
search_domain      = "REPLACE ME" # Provide the name of the network search domain.
vsphere_cluster    = "REPLACE ME" # Provide the cluster name for the machine pool as it appears in vSphere.
datastore_name     = "REPLACE ME" # Provide the datastore name for the machine pool as it appears in vSphere.
network_name       = "REPLACE ME" # Provide the network name for the machine pool as it appears in vSphere.
resource_pool_name = "REPLACE ME" # Provide the resource pool name for the machine pool as it appears in vSphere.
ssh_key            = ""           # Provide the path to your public SSH key. If not provided, a new key pair will be created.
ssh_key_private    = ""           # Provide the path to your private SSH key. If not provided, a new key pair will be created.
