package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterAzure(cluster *models.V1SpectroAzureClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersAzureCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersAzureCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolAzure(cloudConfigId string, machinePool *models.V1AzureMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsAzureMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsAzureMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolAzure(cloudConfigId string, machinePool *models.V1AzureMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsAzureMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsAzureMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolAzure(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsAzureMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsAzureMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1Client) CreateCloudAccountAzure(account *models.V1AzureAccount) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1CloudAccountsAzureCreateParamsWithContext(h.Ctx).WithBody(account)
	success, err := client.V1CloudAccountsAzureCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountAzure(account *models.V1AzureAccount) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1CloudAccountsAzureUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
	_, err = client.V1CloudAccountsAzureUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountAzure(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudAccountsAzureDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1CloudAccountsAzureDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountAzure(uid string) (*models.V1AzureAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsAzureGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsAzureGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsAzure() ([]*models.V1AzureAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsAzureListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsAzureList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1AzureAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1Client) GetCloudConfigAzure(configUID string) (*models.V1AzureCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsAzureGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsAzureGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) ImportClusterAzure(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersAzureImportParamsWithContext(h.Ctx).WithBody(
		&models.V1SpectroAzureClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1SpectroClustersAzureImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}
