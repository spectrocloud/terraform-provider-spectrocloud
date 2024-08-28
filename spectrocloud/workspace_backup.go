package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func toWorkspacePolicies(d *schema.ResourceData) *models.V1WorkspacePolicies {
	policies := toPolicies(d)

	if policies.BackupPolicy == nil {
		return nil
	}

	include_all_clusters, cluster_uids := getExtraFields(d)

	return &models.V1WorkspacePolicies{
		BackupPolicy: &models.V1WorkspaceBackupConfigEntity{
			BackupConfig:       policies.BackupPolicy,
			ClusterUids:        cluster_uids,
			IncludeAllClusters: include_all_clusters,
		},
	}
}

func updateWorkspaceBackupPolicy(c *client.V1Client, d *schema.ResourceData) error {
	if policy := toWorkspaceBackupPolicy(d); policy != nil {
		return c.UpdateWorkspaceBackupConfig(d.Id(), policy)
	}
	return nil
}

func toWorkspaceBackupPolicy(d *schema.ResourceData) *models.V1WorkspaceBackupConfigEntity {
	policy := toBackupPolicy(d)
	if policy == nil {
		return nil
	}

	include_all_clusters, cluster_uids := getExtraFields(d)

	return &models.V1WorkspaceBackupConfigEntity{
		BackupConfig:       policy,
		ClusterUids:        cluster_uids,
		IncludeAllClusters: include_all_clusters,
	}
}

func getExtraFields(d *schema.ResourceData) (bool, []string) {
	include_all_clusters := true
	cluster_uids := make([]string, 0)
	if policies, found := d.GetOk("backup_policy"); found {
		policy := policies.([]interface{})[0].(map[string]interface{})

		if policy["cluster_uids"] != nil {
			for _, uid := range policy["cluster_uids"].(*schema.Set).List() {
				cluster_uids = append(cluster_uids, uid.(string))
			}
		}

		if policy["include_all_clusters"] != nil {
			include_all_clusters = policy["include_all_clusters"].(bool)
		}
	}

	if len(cluster_uids) > 0 {
		return include_all_clusters, cluster_uids
	}

	return include_all_clusters, nil
}

// func detachWorkspaceBackupPolicy(c *client.V1Client) error {
// 	return errors.New("not implemented")
// }
