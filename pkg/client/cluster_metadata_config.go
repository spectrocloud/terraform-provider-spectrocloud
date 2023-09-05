package client

import (
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) UpdateClusterMetadata(uid string, config *models.V1ObjectMetaInputEntitySchema) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1SpectroClustersUIDMetadataUpdateParams().WithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1SpectroClustersUIDMetadataUpdate(params)
	return err
}

func (h *V1Client) UpdateAdditionalClusterMetadata(uid string, additionalMeta *models.V1ClusterMetaAttributeEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}
	params := clusterC.NewV1SpectroClustersUIDClusterMetaAttributeUpdateParams().WithContext(h.Ctx).WithUID(uid).WithBody(additionalMeta)
	_, err = client.V1SpectroClustersUIDClusterMetaAttributeUpdate(params)
	return err
}
