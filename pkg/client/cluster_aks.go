package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)

func (h *V1alpha1Client) CreateClusterAks(cluster *models.V1alpha1SpectroAzureClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersAksCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersAksCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateClusterAks(cluster *models.V1alpha1SpectroAzureClusterEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID
	params := clusterC.NewV1alpha1SpectroClustersAksUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(cluster)
	_, err = client.V1alpha1SpectroClustersAksUpdate(params)
	return err
}

func (h *V1alpha1Client) CreateMachinePoolAks(cloudConfigId string, machinePool *models.V1alpha1AzureMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAksMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsAksMachinePoolCreate(params)
	return err
}

func (h *V1alpha1Client) UpdateMachinePoolAks(cloudConfigId string, machinePool *models.V1alpha1AzureMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAksMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsAksMachinePoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteMachinePoolAks(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAksMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsAksMachinePoolDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudConfigAks(configUID string) (*models.V1alpha1AzureCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsAksGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsAksGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
