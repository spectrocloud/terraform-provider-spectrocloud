package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
)

func toClusterRefs(d *schema.ResourceData) []*models.V1WorkspaceClusterRef {
	clusterRefs := make([]*models.V1WorkspaceClusterRef, 0)

	clusters := d.Get("clusters")
	if clusters == nil {
		return nil
	}
	for _, cluster := range clusters.(*schema.Set).List() {
		clusterValue := cluster.(map[string]interface{})
		uid := clusterValue["uid"].(string)
		clusterRefs = append(clusterRefs, &models.V1WorkspaceClusterRef{
			ClusterUID: uid,
		})
	}

	return clusterRefs
}
