package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterMaas(cluster *models.V1SpectroMaasClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersMaasCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersMaasCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolMaas(cloudConfigId string, machinePool *models.V1MaasMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsMaasMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsMaasMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolMaas(cloudConfigId string, machinePool *models.V1MaasMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsMaasMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsMaasMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolMaas(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsMaasMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsMaasMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1Client) CreateCloudAccountMaas(account *models.V1MaasAccount) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1CloudAccountsMaasCreateParamsWithContext(h.Ctx).WithBody(account)
	success, err := client.V1CloudAccountsMaasCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountMaas(account *models.V1MaasAccount) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1CloudAccountsMaasUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
	_, err = client.V1CloudAccountsMaasUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountMaas(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudAccountsMaasDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1CloudAccountsMaasDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountMaas(uid string) (*models.V1MaasAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsMaasGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsMaasGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsMaas() ([]*models.V1MaasAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsMaasListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsMaasList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1MaasAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1Client) GetCloudConfigMaas(configUID string) (*models.V1MaasCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsMaasGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsMaasGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
