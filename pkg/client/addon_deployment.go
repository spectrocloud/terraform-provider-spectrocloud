package client

import (
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateOrUpdateAddonDeployment(uid string, body *models.V1SpectroClusterProfiles) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	resolveNotification := true
	params := clusterC.NewV1SpectroClustersPatchProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body).WithResolveNotification(&resolveNotification)
	_, err = client.V1SpectroClustersPatchProfiles(params)
	return err
}

func (h *V1Client) DeleteAddonDeploymentValues(uid string, body *models.V1SpectroClusterProfilesDeleteEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1SpectroClustersDeleteProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body)
	_, err = client.V1SpectroClustersDeleteProfiles(params)
	return err
}
