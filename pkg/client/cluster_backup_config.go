package client

import (
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterBackupConfig(uid string) (*models.V1ClusterBackup, error) {
	if h.GetClusterBackupConfigFn != nil {
		return h.GetClusterBackupConfigFn(uid)
	}
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterFeatureBackupGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1ClusterFeatureBackupGet(params)
	if err != nil {
		if herr.IsNotFound(err) || herr.IsBackupNotConfigured(err) {
			return nil, nil
		}
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) CreateClusterBackupConfig(uid string, config *models.V1ClusterBackupConfig) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureBackupCreateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1ClusterFeatureBackupCreate(params)
	return err
}

func (h *V1Client) UpdateClusterBackupConfig(uid string, config *models.V1ClusterBackupConfig) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureBackupUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1ClusterFeatureBackupUpdate(params)
	return err
}

func (h *V1Client) ApplyClusterBackupConfig(uid string, config *models.V1ClusterBackupConfig) error {
	if policy, err := h.GetClusterBackupConfig(uid); err != nil {
		return err
	} else if policy == nil {
		return h.CreateClusterBackupConfig(uid, config)
	} else {
		return h.UpdateClusterBackupConfig(uid, config)
	}
}
