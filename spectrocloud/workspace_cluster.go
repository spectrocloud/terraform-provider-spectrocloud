package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
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
		name := clusterValue["name"].(string)
		clusterRefs = append(clusterRefs, &models.V1WorkspaceClusterRef{
			ClusterName: name,
			ClusterUID:  uid,
		})
	}

	return clusterRefs
}

func toWorkspacePolicies(d *schema.ResourceData) *models.V1WorkspacePolicies {
	policies := toPolicies(d)

	if policies.BackupPolicy == nil {
		return nil
	}

	return &models.V1WorkspacePolicies{
		BackupPolicy: &models.V1WorkspaceBackupConfigEntity{
			BackupConfig:       policies.BackupPolicy,
			ClusterUids:        nil,
			IncludeAllClusters: false,
		},
	}
}
