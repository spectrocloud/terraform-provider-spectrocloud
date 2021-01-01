package client

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	"strings"

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

func (h *V1alpha1Client) GetClusterProfiles() ([]*models.V1alpha1ClusterProfile, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1ClusterProfilesListParamsWithContext(h.ctx)
	response, err := client.V1alpha1ClusterProfilesList(params)
	if err != nil {
		return nil, err
	}

	profiles := make([]*models.V1alpha1ClusterProfile, len(response.Payload.Items))
	for i, profile := range response.Payload.Items {
		profiles[i] = profile
	}

	return profiles, nil
}

func (h *V1alpha1Client) GetPacks(filters []string) ([]*models.V1alpha1PackSummary, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1PacksSummaryListParamsWithContext(h.ctx)
	if filters != nil {
		filterString := ptr.StringPtr(strings.Join(filters,"AND"))
		params = params.WithFilters(filterString)
	}

	response, err := client.V1alpha1PacksSummaryList(params)
	if err != nil {
		return nil, err
	}

	packs := make([]*models.V1alpha1PackSummary, len(response.Payload.Items))
	for i, pack := range response.Payload.Items {
		packs[i] = pack
	}

	return packs, nil
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

