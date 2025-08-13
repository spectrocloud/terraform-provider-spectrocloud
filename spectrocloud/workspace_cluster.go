package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func toClusterRefs(d *schema.ResourceData, c *client.V1Client) []*models.V1WorkspaceClusterRef {
	clusterRefs := make([]*models.V1WorkspaceClusterRef, 0)

	clusters := d.Get("clusters")
	if clusters == nil {
		return nil
	}
	for _, cluster := range clusters.(*schema.Set).List() {
		clusterValue := cluster.(map[string]interface{})
		uid := clusterValue["uid"].(string)

		// Try to get cluster name from the data first (if available from computed field)
		clusterName := ""
		if nameValue, exists := clusterValue["cluster_name"]; exists && nameValue != nil {
			clusterName = nameValue.(string)
		}

		// If cluster name is not available in data and client is provided, fetch it from API
		if clusterName == "" && c != nil {
			if clusterDetails, err := c.GetCluster(uid); err == nil && clusterDetails != nil {
				clusterName = clusterDetails.Metadata.Name
			}
		}

		clusterRefs = append(clusterRefs, &models.V1WorkspaceClusterRef{
			ClusterUID:  uid,
			ClusterName: clusterName,
		})
	}

	return clusterRefs
}
