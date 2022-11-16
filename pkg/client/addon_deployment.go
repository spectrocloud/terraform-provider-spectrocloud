package client

import (
	"github.com/spectrocloud/hapi/models"
	"log"
	"math/rand"
	"time"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) UpdateAddonDeployment(cluster *models.V1SpectroCluster, body *models.V1SpectroClusterProfiles, newProfile *models.V1ClusterProfile) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID

	// check if profile id is the same - update, otherwise, delete and create
	is, replaceUID := isProfileAttachedByName(cluster, newProfile)
	if is {
		body.Profiles[0].ReplaceWithProfile = replaceUID
	}

	resolveNotification := true
	params := clusterC.NewV1SpectroClustersPatchProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body).WithResolveNotification(&resolveNotification)
	err = patchWithRetry(h, client, params)
	return err
}

func isProfileAttachedByName(cluster *models.V1SpectroCluster, newProfile *models.V1ClusterProfile) (bool, string) {

	for _, profile := range cluster.Spec.ClusterProfileTemplates {
		if profile.Name == newProfile.Metadata.Name {
			return true, profile.UID
		}
	}

	return false, ""
}

func (h *V1Client) CreateAddonDeployment(uid string, body *models.V1SpectroClusterProfiles) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	resolveNotification := false // during initial creation we never need to resolve packs.
	params := clusterC.NewV1SpectroClustersPatchProfilesParamsWithContext(h.Ctx).WithUID(uid).WithBody(body).WithResolveNotification(&resolveNotification)
	err = patchWithRetry(h, client, params)
	return err
}

func patchWithRetry(h *V1Client, client clusterC.ClientService, params *clusterC.V1SpectroClustersPatchProfilesParams) error {
	var err error
	log.Printf("Retries: %d", h.retryAttempts)
	for attempt := 0; attempt < h.retryAttempts; attempt++ {
		// small jitter to prevent simultaneous retries
		rand.Seed(time.Now().UnixNano())
		s := rand.Intn(h.retryAttempts) // n will be between 0 and number of retries
		log.Printf("Sleeping %d seconds, retry: %d, cluster:%s, profile:%s, ", s, attempt, params.UID, params.Body.Profiles[0].UID)
		time.Sleep(time.Duration(s) * time.Second)
		_, err = client.V1SpectroClustersPatchProfiles(params)
		if err == nil {
			break
		}
	}
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
