# Terraform configuration
/*import {
  id = "652ce1d0a7296c6b3184555f:project"
  to = spectrocloud_cluster_edge_native.my_cluster
}

import {
  id = "65202c68d160c64b49a34985:tenant"
  to = spectrocloud_cluster_aks.my_cluster
}*/

#Import block for generating configuration for cluster profiles
import {
  id = "64202d68d760c64e49c34682"
  to = spectrocloud_cluster_profile.my_profile
}