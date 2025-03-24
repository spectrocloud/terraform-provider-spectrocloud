package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func toWorkspaceRBACs(d *schema.ResourceData) []*models.V1ClusterRbac {
	rbacs := toClusterRBACsInputEntities(d)
	workspace_rbacs := make([]*models.V1ClusterRbac, 0)
	for _, rbac := range rbacs {
		workspace_rbacs = append(workspace_rbacs,
			&models.V1ClusterRbac{
				Spec: rbac.Spec,
			})
	}

	return workspace_rbacs
}

func toQuota(d *schema.ResourceData) *models.V1WorkspaceQuota {
	wsQuota, ok := d.GetOk("workspace_quota")
	if !ok || len(wsQuota.([]interface{})) == 0 {
		return &models.V1WorkspaceQuota{
			ResourceAllocation: &models.V1WorkspaceResourceAllocation{
				CPUCores:  0,
				MemoryMiB: 0,
			},
		}
	}

	q := wsQuota.([]interface{})[0].(map[string]interface{})
	return &models.V1WorkspaceQuota{
		ResourceAllocation: &models.V1WorkspaceResourceAllocation{
			CPUCores:  float64(q["cpu"].(int)),
			MemoryMiB: float64(q["memory"].(int)),
		},
	}
}
