package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)


func (h *V1alpha1Client) DeleteCluster(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1SpectroClustersDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1SpectroClustersDelete(params)
	return err
}

func (h *V1alpha1Client) GetCluster(uid string) (azureAccount *models.V1alpha1SpectroCluster, err error) {
	client, err := h.getClusterClient()
	if err != nil {
		return
	}

	params := clusterC.NewV1alpha1SpectroClustersGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1SpectroClustersGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

