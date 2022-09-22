package client

import (
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) UpdateAddonDeployment(cluster *models.V1SpectroCluster, body *models.V1SpectroClusterProfiles, newProfile *models.V1ClusterProfile) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID

	// check if profile id is the same - update, otherwise, delete and create
	if isProfileAttachedByName(cluster, newProfile) {
		profile_uids := make([]string, 0)
		profile_uids = append(profile_uids, body.Profiles[0].UID)
		err = h.DeleteAddonDeployment(uid, &models.V1SpectroClusterProfilesDeleteEntity{
			ProfileUids: profile_uids,
		})
		if err != nil {
			return err
		}
	}

	resolveNotification := true
	params := clusterC.NewV1SpectroClustersPatchProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body).WithResolveNotification(&resolveNotification)
	_, err = client.V1SpectroClustersPatchProfiles(params)
	return err
}

func isProfileAttachedByName(cluster *models.V1SpectroCluster, newProfile *models.V1ClusterProfile) bool {

	for _, profile := range cluster.Spec.ClusterProfileTemplates {
		if profile.Name == newProfile.Metadata.Name {
			return true
		}
	}

	return false
}

func (h *V1Client) CreateAddonDeployment(uid string, body *models.V1SpectroClusterProfiles) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	resolveNotification := true
	params := clusterC.NewV1SpectroClustersPatchProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body).WithResolveNotification(&resolveNotification)
	_, err = client.V1SpectroClustersPatchProfiles(params)
	return err
}

func (h *V1Client) DeleteAddonDeployment(uid string, body *models.V1SpectroClusterProfilesDeleteEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1SpectroClustersDeleteProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body)
	_, err = client.V1SpectroClustersDeleteProfiles(params)
	return err
}
