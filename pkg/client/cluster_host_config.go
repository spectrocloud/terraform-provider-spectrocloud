package client

import (
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterHostConfig(uid string) (*models.V1ClusterComplianceScan, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterFeatureComplianceScanGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1ClusterFeatureComplianceScanGet(params)
	if err != nil {
		if herr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return success.Payload, nil
}

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
	if policy, err := h.GetClusterHostConfig(uid); err != nil {
		return err
	} else if policy == nil {
		return h.UpdateClusterHostConfig(uid, config)
	} else {
		return h.UpdateClusterHostConfig(uid, config)
	}
}
