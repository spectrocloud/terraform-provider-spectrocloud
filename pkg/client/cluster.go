package client

import (
	"strings"

	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) DeleteCluster(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1SpectroClustersDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1SpectroClustersDelete(params)
	return err
}

func (h *V1Client) GetCluster(uid string) (*models.V1SpectroCluster, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1SpectroClustersGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1SpectroClustersGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// special check if the cluster is marked deleted
	cluster := success.Payload
	if cluster.Status.State == "Deleted" {
		return nil, nil
	}

	return success.Payload, nil
}

func (h *V1Client) GetClusterKubeConfig(uid string) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	builder := new(strings.Builder)
	params := clusterC.NewV1SpectroClustersUIDKubeConfigParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1SpectroClustersUIDKubeConfig(params, builder)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (h *V1Client) GetClusterImportManifest(uid string) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	builder := new(strings.Builder)
	params := clusterC.NewV1SpectroClustersUIDInstallerManifestParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1SpectroClustersUIDInstallerManifest(params, builder)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (h *V1Client) UpdateClusterProfileValues(uid string, profiles *models.V1SpectroClusterProfiles) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	resolveNotification := true
	params := clusterC.NewV1SpectroClustersUpdateProfilesParamsWithContext(h.ctx).WithUID(uid).
		WithBody(profiles).WithResolveNotification(&resolveNotification)
	_, err = client.V1SpectroClustersUpdateProfiles(params)
	return err
}

func (h *V1Client) GetClusterBackupConfig(uid string) (*models.V1ClusterBackup, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterFeatureBackupGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1ClusterFeatureBackupGet(params)
	if err != nil {
		if herr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) CreateClusterBackupConfig(uid string, config *models.V1ClusterBackupConfig) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureBackupCreateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
	_, err = client.V1ClusterFeatureBackupCreate(params)
	return err
}

func (h *V1Client) UpdateClusterBackupConfig(uid string, config *models.V1ClusterBackupConfig) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureBackupUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
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

func (h *V1Client) GetClusterScanConfig(uid string) (*models.V1ClusterComplianceScan, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterFeatureComplianceScanGetParamsWithContext(h.ctx).WithUID(uid)
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
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureComplianceScanCreateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
	_, err = client.V1ClusterFeatureComplianceScanCreate(params)
	return err
}

func (h *V1Client) UpdateClusterScanConfig(uid string, config *models.V1ClusterComplianceScheduleConfig) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterFeatureComplianceScanUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
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
