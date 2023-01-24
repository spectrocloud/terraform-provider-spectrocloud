package client

import (
	"errors"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterGroup(cluster *models.V1ClusterGroupEntity) (string, error) {
	if h.CreateClusterGroupFn != nil {
		return h.CreateClusterGroupFn(cluster)
	}

	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1ClusterGroupsCreateParams().WithBody(cluster)
	success, err := client.V1ClusterGroupsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) DeleteClusterGroup(uid string) error {
	if h.DeleteClusterGroupFn != nil {
		return h.DeleteClusterGroupFn(uid)
	}
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterGroupsUIDDeleteParams().WithUID(uid)
	_, err = client.V1ClusterGroupsUIDDelete(params)
	return err
}

func (h *V1Client) GetClusterGroup(uid string) (*models.V1ClusterGroup, error) {
	if h.GetClusterGroupFn != nil {
		return h.GetClusterGroupFn(uid)
	}

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

	params := clusterC.NewV1ClusterGroupsUIDGetParams().WithUID(uid)
	success, err := client.V1ClusterGroupsUIDGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	cluster := success.Payload
	return cluster, nil
}

func (h *V1Client) GetClusterGroupByName(name string, ClusterGroupContext string) (*models.V1ObjectScopeEntity, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	var params *clusterC.V1ClusterGroupsHostClusterMetadataParams
	switch ClusterGroupContext {
	case "system":
		params = clusterC.NewV1ClusterGroupsHostClusterMetadataParams()
		break
	case "tenant":
		params = clusterC.NewV1ClusterGroupsHostClusterMetadataParams().WithContext(h.Ctx)
		break
	default:
		return nil, errors.New("invalid scope")
	}

	success, err := client.V1ClusterGroupsHostClusterMetadata(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for _, group := range success.Payload.Items {
		if group.Name == name && group.Scope == ClusterGroupContext { // tenant or system. keep it to extend to project in future.
			return group, nil
		}
	}

	return nil, nil
}

// Update cluster group metadata by invoking V1ClusterGroupsUIDMetaUpdate hapi api
func (h *V1Client) UpdateClusterGroupMeta(clusterGroup *models.V1ClusterGroupEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterGroupsUIDMetaUpdateParams().WithUID(clusterGroup.Metadata.UID)
	params = params.WithBody(&models.V1ObjectMeta{
		Name:        clusterGroup.Metadata.Name,
		Labels:      clusterGroup.Metadata.Labels,
		Annotations: clusterGroup.Metadata.Annotations,
	})
	_, err = client.V1ClusterGroupsUIDMetaUpdate(params)
	return err
}

// Update cluster group by invoking V1ClusterGroupsUIDHostClusterUpdate hapi api
func (h *V1Client) UpdateClusterGroup(uid string, clusterGroup *models.V1ClusterGroupHostClusterEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterGroupsUIDHostClusterUpdateParams().WithUID(uid)
	params = params.WithBody(clusterGroup)
	_, err = client.V1ClusterGroupsUIDHostClusterUpdate(params)
	return err
}
