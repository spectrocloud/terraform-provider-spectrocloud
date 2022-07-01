package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
)

func toWorkspaceRBACs(d *schema.ResourceData) []*models.V1ClusterRbac {
	rbacs := toClusterRBACs(d)
	workspace_rbacs := make([]*models.V1ClusterRbac, 0)
	for _, rbac := range rbacs {
		workspace_rbacs = append(workspace_rbacs,
			&models.V1ClusterRbac{
				Spec: rbac.Spec,
			})
	}

	return workspace_rbacs
}

func toUpdateWorkspaceRBACs(d *schema.ResourceData) *models.V1ClusterRbac {
	rbacs := toWorkspaceRBACs(d)
	bindings := make([]*models.V1ClusterRbacBinding, 0)
	for _, rbac := range rbacs {
		for _, binding := range rbac.Spec.Bindings {
			bindings = append(bindings, binding)
		}
	}

	return &models.V1ClusterRbac{Spec: &models.V1ClusterRbacSpec{Bindings: bindings}}
}

func toQuota(d *schema.ResourceData) *models.V1WorkspaceQuota {
	return &models.V1WorkspaceQuota{
		ResourceAllocation: &models.V1WorkspaceResourceAllocation{
			CPUCores:  0,
			MemoryMiB: 0,
		},
	}
}
