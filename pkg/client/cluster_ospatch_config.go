package client

import (
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) UpdateClusterOsPatchConfig(uid string, config *models.V1OsPatchEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1SpectroClustersUIDOsPatchUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
	_, err = client.V1SpectroClustersUIDOsPatchUpdate(params)
	return err
}
