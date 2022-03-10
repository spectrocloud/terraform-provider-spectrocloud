package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterLibvirt(cluster *models.V1SpectroLibvirtClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersLibvirtCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersLibvirtCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolLibvirt(cloudConfigId string, machinePool *models.V1LibvirtMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsLibvirtMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsLibvirtMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolLibvirt(cloudConfigId string, machinePool *models.V1LibvirtMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsLibvirtMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsLibvirtMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolLibvirt(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsLibvirtMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsLibvirtMachinePoolDelete(params)
	return err
}

func (h *V1Client) GetCloudConfigLibvirt(configUID string) (*models.V1LibvirtCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsLibvirtGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsLibvirtGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
