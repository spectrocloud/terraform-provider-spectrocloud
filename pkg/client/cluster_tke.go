package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterTke(cluster *models.V1SpectroTencentClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersTkeCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersTkeCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolTke(cloudConfigId string, machinePool *models.V1TencentMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsTkeMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsTkeMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolTke(cloudConfigId string, machinePool *models.V1TencentMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsTkeMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsTkeMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolTke(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsTkeMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsTkeMachinePoolDelete(params)
	return err
}

func (h *V1Client) GetCloudConfigTke(configUID string) (*models.V1TencentCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsTkeGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsTkeGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) CreateCloudAccountTke(account *models.V1TencentAccount) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1CloudAccountsTencentCreateParamsWithContext(h.Ctx).WithBody(account)
	success, err := client.V1CloudAccountsTencentCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountTencent(account *models.V1TencentAccount) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1CloudAccountsTencentUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
	_, err = client.V1CloudAccountsTencentUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountTke(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudAccountsTencentDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1CloudAccountsTencentDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountTke(uid string) (*models.V1TencentAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsTencentGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsTencentGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsTke() ([]*models.V1TencentAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsTencentListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsTencentList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1TencentAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}
