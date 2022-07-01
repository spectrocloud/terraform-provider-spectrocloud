package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	hashboardC "github.com/spectrocloud/hapi/hashboard/client/v1"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateWorkspace(workspace *models.V1WorkspaceEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1WorkspacesCreateParamsWithContext(h.Ctx).WithBody(workspace)
	success, err := client.V1WorkspacesCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) GetWorkspace(uid string) (*models.V1Workspace, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1WorkspacesUIDGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1WorkspacesUIDGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	workspace := success.Payload

	return workspace, nil
}

func (h *V1Client) GetWorkspaceByName(name string) (*models.V1DashboardWorkspace, error) {
	client, err := h.GetHashboard()
	if err != nil {
		return nil, err
	}

	params := hashboardC.NewV1DashboardWorkspacesListParamsWithContext(h.Ctx)
	success, err := client.V1DashboardWorkspacesList(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for _, workspace := range success.Payload.Items {
		if workspace.Metadata.Name == name {
			return workspace, nil
		}
	}

	return nil, nil
}

func (h *V1Client) UpdateWorkspaceResourceAllocation(uid string, wo *models.V1WorkspaceResourceAllocationsEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1WorkspacesUIDResourceAllocationsUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(wo)
	if _, err := client.V1WorkspacesUIDResourceAllocationsUpdate(params); err != nil {
		return err
	}
	return nil
}

func (h *V1Client) UpdateWorkspaceRBACS(uid string, rbac_uid string, wo *models.V1ClusterRbac) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1WorkspacesUIDClusterRbacUpdateParamsWithContext(h.Ctx).WithUID(uid).WithClusterRbacUID(rbac_uid).WithBody(wo)
	if _, err := client.V1WorkspacesUIDClusterRbacUpdate(params); err != nil {
		return err
	}
	return nil
}

func (h *V1Client) DeleteWorkspace(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1WorkspacesUIDDeleteParamsWithContext(h.Ctx).WithUID(uid)
	if _, err := client.V1WorkspacesUIDDelete(params); err != nil {
		return err
	}
	return nil
}
