package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"
)

func (h *V1Client) CreateClusterOpenStack(cluster *models.V1SpectroOpenStackClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersOpenStackCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersOpenStackCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateCloudAccountOpenStack(account *models.V1OpenStackAccount) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1CloudAccountsOpenStackCreateParamsWithContext(h.ctx).WithBody(account)
	success, err := client.V1CloudAccountsOpenStackCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolOpenStack(cloudConfigId string, machinePool *models.V1OpenStackMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsOpenStackMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsOpenStackMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolOpenStack(cloudConfigId string, machinePool *models.V1OpenStackMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsOpenStackMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsOpenStackMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolOpenStack(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudConfigsOpenStackMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsOpenStackMachinePoolDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountOpenStack(uid string) (*models.V1OpenStackAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsOpenStackGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1CloudAccountsOpenStackGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudConfigOpenStack(configUID string) (*models.V1OpenStackCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsOpenStackGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsOpenStackGet(params)

	if herr.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) UpdateCloudAccountOpenStack(account *models.V1OpenStackAccount) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1CloudAccountsOpenStackUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(account)
	_, err = client.V1CloudAccountsOpenStackUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountOpenStack(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1CloudAccountsOpenStackDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1CloudAccountsOpenStackDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountsOpenStack() ([]*models.V1OpenStackAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsOpenStackListParamsWithContext(h.ctx)
	response, err := client.V1CloudAccountsOpenStackList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1OpenStackAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}
