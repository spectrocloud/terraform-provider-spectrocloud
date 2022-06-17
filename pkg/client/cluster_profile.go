package client

import (
	"strings"

	"github.com/spectrocloud/gomi/pkg/ptr"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) DeleteClusterProfile(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1ClusterProfilesDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1ClusterProfilesDelete(params)
	return err
}

func (h *V1Client) GetClusterProfile(uid string) (*models.V1ClusterProfile, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterProfilesGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1ClusterProfilesGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetClusterProfileManifestPack(clusterProfileUID, packName string) ([]*models.V1ManifestEntity, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	//params := clusterC.NewV1ClusterProfilesGetParamsWithContext(h.ctx).WithUID(uid)
	params := clusterC.NewV1ClusterProfilesUIDPacksUIDManifestsParamsWithContext(h.ctx).
		WithUID(clusterProfileUID).WithPackName(packName)
	success, err := client.V1ClusterProfilesUIDPacksUIDManifests(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload.Items, nil
}

func (h *V1Client) GetClusterProfiles() ([]*models.V1ClusterProfile, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	limit := int64(0)
	params := clusterC.NewV1ClusterProfilesListParamsWithContext(h.ctx).WithLimit(&limit)
	response, err := client.V1ClusterProfilesList(params)
	if err != nil {
		return nil, err
	}

	profiles := make([]*models.V1ClusterProfile, len(response.Payload.Items))
	for i, profile := range response.Payload.Items {
		profiles[i] = profile
	}

	return profiles, nil
}

func (h *V1Client) GetPacks(filters []string, registryUID string) ([]*models.V1PackSummary, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1PacksSummaryListParamsWithContext(h.ctx)
	if filters != nil {
		filterString := ptr.StringPtr(strings.Join(filters, "AND"))
		params = params.WithFilters(filterString)
	}

	response, err := client.V1PacksSummaryList(params)
	if err != nil {
		return nil, err
	}

	packs := make([]*models.V1PackSummary, 0)
	for _, pack := range response.Payload.Items {
		if registryUID == "" || pack.Spec.RegistryUID == registryUID {
			packs = append(packs, pack)
		}
	}

	return packs, nil
}

func (h *V1Client) GetPack(uid string) (*models.V1PackTagEntity, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1PacksUIDParamsWithContext(h.ctx).WithUID(uid)
	response, err := client.V1PacksUID(params)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (h *V1Client) GetPackRegistry(pack *models.V1PackRef) string {
	if pack.PackUID == "uid" || pack.Type == "manifest" {
		registry, err := h.GetPackRegistryCommonByName("Public Repo")
		if err != nil {
			return ""
		}
		return registry.UID
	}

	PackTagEntity, err := h.GetPack(pack.PackUID)
	if err != nil {
		return ""
	}

	return PackTagEntity.RegistryUID
}

func (h *V1Client) UpdateClusterProfile(clusterProfile *models.V1ClusterProfileUpdateEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := clusterProfile.Metadata.UID
	params := clusterC.NewV1ClusterProfilesUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(clusterProfile)
	_, err = client.V1ClusterProfilesUpdate(params)
	return err
}

func (h *V1Client) CreateClusterProfile(clusterProfile *models.V1ClusterProfileEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1ClusterProfilesCreateParamsWithContext(h.ctx).WithBody(clusterProfile)
	success, err := client.V1ClusterProfilesCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) PublishClusterProfile(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1ClusterProfilesPublishParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1ClusterProfilesPublish(params)
	if err != nil {
		return err
	}

	return nil
}
