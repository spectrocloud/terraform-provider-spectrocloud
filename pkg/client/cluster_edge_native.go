package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterEdgeNative(cluster *models.V1SpectroEdgeNativeClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersEdgeNativeCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersEdgeNativeCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolEdgeNative(cloudConfigId string, machinePool *models.V1EdgeNativeMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsEdgeNativeMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsEdgeNativeMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolEdgeNative(cloudConfigId string, machinePool *models.V1EdgeNativeMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsEdgeNativeMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsEdgeNativeMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolEdgeNative(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsEdgeNativeMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsEdgeNativeMachinePoolDelete(params)
	return err
}

func (h *V1Client) GetCloudConfigEdgeNative(configUID string) (*models.V1EdgeNativeCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsEdgeNativeGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsEdgeNativeGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
