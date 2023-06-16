package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ForceDeleteTimeoutSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "5m",
		Description: `A deletion timeout value that triggers the cluster to be force deleted. 
If a cluster is stuck in the Deleting state for a minimum of 15 minutes, it becomes eligible for force deletion. 
You can force delete a cluster from the tenant and project admin scope. 
A force deletion can result in Palette-provisioned resources being missed in the removal process. 
Verify there are no remaining resources. 
Refer to the [Cluster Removal](https://docs.spectrocloud.com/clusters/cluster-management/remove-clusters#forcedeleteacluster) 
reference page to learn more about force delete. The default value is ` + "`5m` (5 minutes).",
	}
}
