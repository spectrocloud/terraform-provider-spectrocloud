package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func flattenWorkspaceClusters(workspace *models.V1Workspace) []interface{} {
	clusters := workspace.Spec.ClusterRefs

	if len(clusters) > 0 {
		wsp_clusters := make([]interface{}, 0)

		for _, cluster := range clusters {
			wsp_cluster := make(map[string]interface{})

			wsp_cluster["uid"] = cluster.ClusterUID

			wsp_clusters = append(wsp_clusters, wsp_cluster)
		}

		return wsp_clusters
	} else {
		return make([]interface{}, 0)
	}
}

func flattenWorkspaceBackupPolicy(backup *models.V1WorkspaceBackup, d *schema.ResourceData) []interface{} {
	result := make([]interface{}, 0, 1)
	if backup.Spec.Config == nil && backup.Spec.Config.BackupConfig == nil {
		return result
	}
	result = flattenBackupPolicy(backup.Spec.Config.BackupConfig, d)
	data := result[0].(map[string]interface{})
	data["cluster_uids"] = backup.Spec.Config.ClusterUids
	data["include_all_clusters"] = backup.Spec.Config.IncludeAllClusters
	return result
}
