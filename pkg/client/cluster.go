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

func (h *V1Client) GetClusterNotification(uid string) (bool, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return false, err
	}

	params := clusterC.NewV1SpectroClustersGetProfileUpdatesParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1SpectroClustersGetProfileUpdates(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// special check if the cluster is marked deleted
	payload := success.Payload
	if len(payload.Profiles) > 0 {
		return true, nil
	} else {
		return false, nil
	}
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
		if herr.IsNotFound(err) {
			return "", nil
		}
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
