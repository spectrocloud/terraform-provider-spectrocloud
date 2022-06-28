package client

import (
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterScanConfig(uid string) (*models.V1ClusterComplianceScan, error) {
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

func (h *V1Client) CreateClusterScanConfig(uid string, config *models.V1ClusterComplianceScheduleConfig) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureComplianceScanCreateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1ClusterFeatureComplianceScanCreate(params)
	return err
}

func (h *V1Client) UpdateClusterScanConfig(uid string, config *models.V1ClusterComplianceScheduleConfig) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureComplianceScanUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1ClusterFeatureComplianceScanUpdate(params)
	return err
}

func (h *V1Client) ApplyClusterScanConfig(uid string, config *models.V1ClusterComplianceScheduleConfig) error {
	if policy, err := h.GetClusterScanConfig(uid); err != nil {
		return err
	} else if policy == nil {
		return h.CreateClusterScanConfig(uid, config)
	} else {
		return h.UpdateClusterScanConfig(uid, config)
	}
}
