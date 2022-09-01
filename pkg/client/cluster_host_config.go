package client

import (
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) UpdateClusterHostConfig(uid string, config *models.V1HostClusterConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1HostClusterConfigUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1HostClusterConfigUpdate(params)
	return err
}

func (h *V1Client) ApplyClusterHostConfig(uid string, config *models.V1HostClusterConfigEntity) error {
	if policy, err := h.GetClusterScanConfig(uid); err != nil {
		return err
	} else if policy == nil {
		return h.UpdateClusterHostConfig(uid, config)
	} else {
		return h.UpdateClusterHostConfig(uid, config)
	}
}
