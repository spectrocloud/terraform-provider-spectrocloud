package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetApplication(uid string) (*models.V1AppDeployment, error) {
	if h.GetApplicationFn != nil {
		return h.GetApplicationFn(uid)
	}

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
	application := success.Payload //success.Payload.Spec.Config.Target.ClusterRef.UID
	return application, nil
}

func (h *V1Client) CreateApplicationWithNewSandboxCluster(body *models.V1AppDeploymentClusterGroupEntity) (string, error) {
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

func (h *V1Client) CreateApplicationWithExistingSandboxCluster(body *models.V1AppDeploymentVirtualClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1AppDeploymentsVirtualClusterCreateParams().WithContext(h.Ctx).WithBody(body)
	success, err := client.V1AppDeploymentsVirtualClusterCreate(params)

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
