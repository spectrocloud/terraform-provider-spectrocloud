package client

import (
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterNamespaceConfig(uid string) (*models.V1ClusterNamespaceResources, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1SpectroClustersUIDConfigNamespacesGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1SpectroClustersUIDConfigNamespacesGet(params)
	if err != nil {
		if herr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return success.Payload, nil
}

// no create for namespecase, there is only update.
func (h *V1Client) UpdateClusterNamespaceConfig(uid string, config *models.V1ClusterNamespaceResourcesUpdateEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1SpectroClustersUIDConfigNamespacesUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
	_, err = client.V1SpectroClustersUIDConfigNamespacesUpdate(params)
	return err
}

func (h *V1Client) ApplyClusterNamespaceConfig(uid string, config []*models.V1ClusterNamespaceResourceInputEntity) error {
	if _, err := h.GetClusterNamespaceConfig(uid); err != nil {
		return err
	} else {
		return h.UpdateClusterNamespaceConfig(uid, toUpdateNamespace(config)) // update method is same as create
	}
}

func toUpdateNamespace(config []*models.V1ClusterNamespaceResourceInputEntity) *models.V1ClusterNamespaceResourcesUpdateEntity {
	return &models.V1ClusterNamespaceResourcesUpdateEntity{
		Namespaces: config,
	}
}
