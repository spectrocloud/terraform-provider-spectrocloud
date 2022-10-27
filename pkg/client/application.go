package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetApplication(uid string) (*models.V1AppDeployment, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1AppDeploymentsUIDGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1AppDeploymentsUIDGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// special check if the cluster is marked deleted
	application := success.Payload
	return application, nil
}

func (h *V1Client) UpdateApplication(cluster *models.V1SpectroCluster, body *models.V1SpectroClusterProfiles, newProfile *models.V1ClusterProfile) error {
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

func (h *V1Client) CreateApplication(body *models.V1AppDeploymentClusterGroupEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1AppDeploymentsClusterGroupCreateParams().WithContext(h.Ctx).WithBody(body)
	success, err := client.V1AppDeploymentsClusterGroupCreate(params)

	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) DeleteApplication(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}
	params := clusterC.NewV1AppDeploymentsUIDDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1AppDeploymentsUIDDelete(params)
	return err
}
