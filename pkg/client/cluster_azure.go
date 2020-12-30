package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)


func (h *V1alpha1Client) CreateClusterAzure(cluster *models.V1alpha1SpectroAzureClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersAzureCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersAzureCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateClusterAzure(cluster *models.V1alpha1SpectroAzureClusterEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID
	params := clusterC.NewV1alpha1SpectroClustersAzureUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(cluster)
	_, err = client.V1alpha1SpectroClustersAzureUpdate(params)
	return err
}

func (h *V1alpha1Client) CreateMachinePoolAzure(cloudConfigId string, machinePool *models.V1alpha1AzureMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAzureMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsAzureMachinePoolCreate(params)
	return err
}

func (h *V1alpha1Client) UpdateMachinePoolAzure(cloudConfigId string, machinePool *models.V1alpha1AzureMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAzureMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsAzureMachinePoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteMachinePoolAzure(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAzureMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsAzureMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1alpha1Client) CreateCloudAccountAzure(account *models.V1alpha1AzureAccount) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1CloudAccountsAzureCreateParamsWithContext(h.ctx).WithBody(account)
	success, err := client.V1alpha1CloudAccountsAzureCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateCloudAccountAzure(account *models.V1alpha1AzureAccount) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1alpha1CloudAccountsAzureUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(account)
	_, err = client.V1alpha1CloudAccountsAzureUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteCloudAccountAzure(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudAccountsAzureDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1CloudAccountsAzureDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudAccountAzure(uid string) (*models.V1alpha1AzureAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsAzureGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1CloudAccountsAzureGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) GetCloudAccountsAzure() ([]*models.V1alpha1AzureAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsAzureListParamsWithContext(h.ctx)
	response, err := client.V1alpha1CloudAccountsAzureList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1alpha1AzureAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1alpha1Client) GetCloudConfigAzure(configUID string) (*models.V1alpha1AzureCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsAzureGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsAzureGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
