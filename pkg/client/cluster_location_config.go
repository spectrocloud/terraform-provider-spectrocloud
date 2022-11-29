package client

import (
	"errors"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterLocationConfig(uid string) (*models.V1ClusterLocation, error) {
	if clusterStatus, err := h.GetClusterWithoutStatus(uid); err != nil {
		return nil, err
	} else if clusterStatus != nil && clusterStatus.Status != nil && clusterStatus.Status.Location != nil {
		return clusterStatus.Status.Location, nil
	}

	return nil, errors.New("Error while reading cluster location.")
}

func (h *V1Client) UpdateClusterLocationConfig(uid string, config *models.V1SpectroClusterLocationInputEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1SpectroClustersUIDLocationPutParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1SpectroClustersUIDLocationPut(params)
	return err
}

func (h *V1Client) ApplyClusterLocationConfig(uid string, config *models.V1SpectroClusterLocationInputEntity) error {
	if curentConfig, err := h.GetClusterLocationConfig(uid); err != nil {
		return err
	} else if curentConfig == nil {
		return h.UpdateClusterLocationConfig(uid, config)
	} else {
		return h.UpdateClusterLocationConfig(uid, config)
	}
}
