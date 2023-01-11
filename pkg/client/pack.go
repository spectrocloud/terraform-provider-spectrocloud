package client

import (
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"strings"

	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

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

func (h *V1Client) GetPacksByNameAndRegistry(name string, registryUID string) (*models.V1PackTagEntity, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1PacksNameRegistryUIDListParamsWithContext(h.Ctx).WithPackName(name).WithRegistryUID(registryUID)

	response, err := client.V1PacksNameRegistryUIDList(params)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
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
