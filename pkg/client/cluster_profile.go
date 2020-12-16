package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)


func (h *V1alpha1Client) DeleteClusterProfile(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1ClusterProfilesDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1ClusterProfilesDelete(params)
	return err
}

func (h *V1alpha1Client) GetClusterProfile(uid string) (*models.V1alpha1ClusterProfile, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1ClusterProfilesGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1ClusterProfilesGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) UpdateClusterProfile(clusterProfile *models.V1alpha1ClusterProfileEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := clusterProfile.Metadata.UID
	params := clusterC.NewV1alpha1ClusterProfilesUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(clusterProfile)
	_, err = client.V1alpha1ClusterProfilesUpdate(params)
	return err
}

func (h *V1alpha1Client) CreateClusterProfile(cluster *models.V1alpha1ClusterProfileEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1ClusterProfilesCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1ClusterProfilesCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) PublishClusterProfile(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1alpha1ClusterProfilesPublishParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1ClusterProfilesPublish(params)
	if err != nil {
		return err
	}

	return nil
}

