package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) DeleteClusterGroup(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1ClusterGroupsUIDDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1ClusterGroupsUIDDelete(params)
	return err
}

func (h *V1Client) GetClusterGroup(uid string) (*models.V1ClusterGroup, error) {
	group, err := h.GetClusterGroupWithoutStatus(uid)
	if err != nil {
		return nil, err
	}

	if group == nil && group.Status.IsActive {
		return nil, nil
	}

	return group, nil
}

func (h *V1Client) GetClusterGroupWithoutStatus(uid string) (*models.V1ClusterGroup, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterGroupsUIDGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1ClusterGroupsUIDGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// special check if the cluster is marked deleted
	cluster := success.Payload
	return cluster, nil
}

func (h *V1Client) GetClusterGroupByName(name string, scope string) (*models.V1ObjectScopeEntity, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterGroupsHostClusterMetadataParams().WithContext(h.Ctx)
	success, err := client.V1ClusterGroupsHostClusterMetadata(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for _, group := range success.Payload.Items {
		if group.Name == name && group.Scope == scope { // tenant or system
			return group, nil
		}
	}

	return nil, nil
}
