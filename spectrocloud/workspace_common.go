package spectrocloud

import "github.com/spectrocloud/hapi/models"

func flattenWorkspaceClusters(workspace *models.V1Workspace) []interface{} {
	clusters := workspace.Spec.ClusterRefs

	if len(clusters) > 0 {
		wsp_clusters := make([]interface{}, 0)

		for _, cluster := range clusters {
			wsp_cluster := make(map[string]interface{})

			wsp_cluster["name"] = cluster.ClusterName
			wsp_cluster["uid"] = cluster.ClusterUID

			wsp_clusters = append(wsp_clusters, wsp_cluster)
		}

		return wsp_clusters
	} else {
		return make([]interface{}, 0)
	}
}

func flattenWorkspaceBackupPolicy(workspace *models.V1Workspace) []interface{} {
	result := make([]interface{}, 0, 1)
	data := make(map[string]interface{})
	if workspace.Spec.Policies == nil || workspace.Spec.Policies.BackupPolicy == nil {
		return result
	}
	backupConfig := workspace.Spec.Policies.BackupPolicy.BackupConfig
	data["schedule"] = backupConfig.Schedule
	data["backup_location_id"] = backupConfig.BackupLocationUID
	data["prefix"] = backupConfig.BackupPrefix
	data["namespaces"] = backupConfig.Namespaces
	data["expiry_in_hour"] = backupConfig.DurationInHours
	data["include_disks"] = backupConfig.IncludeAllDisks
	data["include_cluster_resources"] = backupConfig.IncludeClusterResources
	result = append(result, data)
	return result
}
