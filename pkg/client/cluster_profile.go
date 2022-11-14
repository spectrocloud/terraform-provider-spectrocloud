package client

import (
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"strings"

	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) DeleteClusterProfile(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	profile, err := h.GetClusterProfile(uid)
	if err != nil {
		return nil
	}

	var params *clusterC.V1ClusterProfilesDeleteParams
	switch profile.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1ClusterProfilesDeleteParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesDeleteParams().WithUID(uid)
		break
	default:
		break
	}

	_, err = client.V1ClusterProfilesDelete(params)
	return err
}

func (h *V1Client) GetClusterProfile(uid string) (*models.V1ClusterProfile, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1ClusterProfilesGetParamsWithContext(h.Ctx).WithUID(uid)
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
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	//params := clusterC.NewV1ClusterProfilesGetParamsWithContext(h.ctx).WithUID(uid)
	params := clusterC.NewV1ClusterProfilesUIDPacksUIDManifestsParamsWithContext(h.Ctx).
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
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	limit := int64(0)
	params := clusterC.NewV1ClusterProfilesListParamsWithContext(h.Ctx).WithLimit(&limit)
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
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1PacksSummaryListParamsWithContext(h.Ctx)
	if filters != nil {
		filterString := types.Ptr(strings.Join(filters, "AND"))
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
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1PacksUIDParamsWithContext(h.Ctx).WithUID(uid)
	response, err := client.V1PacksUID(params)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (h *V1Client) GetPackRegistry(packUID string, packType string) string {
	if packUID == "uid" || packType == "manifest" {
		registry, err := h.GetPackRegistryCommonByName("Public Repo")
		if err != nil {
			return ""
		}
		return registry.UID
	}

	PackTagEntity, err := h.GetPack(packUID)
	if err != nil {
		return ""
	}

	return PackTagEntity.RegistryUID
}

func (h *V1Client) PatchClusterProfile(clusterProfile *models.V1ClusterProfileUpdateEntity, metadata *models.V1ProfileMetaEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	uid := clusterProfile.Metadata.UID
	var params *clusterC.V1ClusterProfilesUIDMetadataUpdateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesUIDMetadataUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(metadata)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesUIDMetadataUpdateParams().WithUID(uid).WithBody(metadata)
		break
	default:
		break
	}
	_, err = client.V1ClusterProfilesUIDMetadataUpdate(params)
	return err
}

func (h *V1Client) UpdateClusterProfile(clusterProfile *models.V1ClusterProfileUpdateEntity, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	uid := clusterProfile.Metadata.UID
	var params *clusterC.V1ClusterProfilesUpdateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(clusterProfile)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesUpdateParams().WithUID(uid).WithBody(clusterProfile)
		break
	default:
		break
	}
	_, err = client.V1ClusterProfilesUpdate(params)
	return err
}

func (h *V1Client) CreateClusterProfile(clusterProfile *models.V1ClusterProfileEntity, ProfileContext string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	var params *clusterC.V1ClusterProfilesCreateParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesCreateParamsWithContext(h.Ctx).WithBody(clusterProfile)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesCreateParams().WithBody(clusterProfile)
		break
	default:
		break
	}
	success, err := client.V1ClusterProfilesCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) PublishClusterProfile(uid string, ProfileContext string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	var params *clusterC.V1ClusterProfilesPublishParams
	switch ProfileContext {
	case "project":
		params = clusterC.NewV1ClusterProfilesPublishParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1ClusterProfilesPublishParams().WithUID(uid)
		break
	default:
		break
	}
	_, err = client.V1ClusterProfilesPublish(params)
	if err != nil {
		return err
	}

	return nil
}
