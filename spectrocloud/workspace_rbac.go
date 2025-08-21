package spectrocloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/constants"
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

func toQuota(d *schema.ResourceData) (*models.V1WorkspaceQuota, error) {
	wsQuota, ok := d.GetOk("workspace_quota")
	if !ok || len(wsQuota.([]interface{})) == 0 {
		return &models.V1WorkspaceQuota{
			ResourceAllocation: &models.V1WorkspaceResourceAllocation{
				CPUCores:  0,
				MemoryMiB: 0,
			},
		}, nil
	}

	q := wsQuota.([]interface{})[0].(map[string]interface{})
	resourceAllocation := &models.V1WorkspaceResourceAllocation{
		CPUCores:  float64(q["cpu"].(int)),
		MemoryMiB: float64(q["memory"].(int)),
	}

	// Handle GPU configuration if specified
	if gpuVal, exists := q["gpu"]; exists && gpuVal.(int) > 0 {
		gpuInt := gpuVal.(int)
		if gpuInt > constants.Int32MaxValue {
			return nil, fmt.Errorf("gpu value %d is out of range for int32", gpuInt)
		}
		provider := "nvidia" // Default to nvidia as it's the only supported provider
		resourceAllocation.GpuConfig = &models.V1GpuConfig{
			Limit:    SafeInt32(gpuInt),
			Provider: &provider,
		}
	}

	return &models.V1WorkspaceQuota{
		ResourceAllocation: resourceAllocation,
	}, nil
}
