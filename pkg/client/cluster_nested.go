package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterNested(cluster *models.V1SpectroNestedClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersNestedCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersNestedCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolNested(cloudConfigId string, machinePool *models.V1NestedMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsNestedMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsNestedMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolNested(cloudConfigId string, machinePool *models.V1NestedMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsNestedMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsNestedMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolNested(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsNestedMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsNestedMachinePoolDelete(params)
	return err
}

func (h *V1Client) GetCloudConfigNested(configUID string) (*models.V1NestedCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsNestedGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsNestedGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
