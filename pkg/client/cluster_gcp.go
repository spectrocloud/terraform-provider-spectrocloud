package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterGcp(cluster *models.V1SpectroGcpClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersGcpCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersGcpCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolGcp(cloudConfigId string, machinePool *models.V1GcpMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsGcpMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsGcpMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolGcp(cloudConfigId string, machinePool *models.V1GcpMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsGcpMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsGcpMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolGcp(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsGcpMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsGcpMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1Client) CreateCloudAccountGcp(account *models.V1GcpAccountEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1CloudAccountsGcpCreateParamsWithContext(h.Ctx).WithBody(account)
	success, err := client.V1CloudAccountsGcpCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountGcp(account *models.V1GcpAccountEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1CloudAccountsGcpUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
	_, err = client.V1CloudAccountsGcpUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountGcp(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudAccountsGcpDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1CloudAccountsGcpDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountGcp(uid string) (*models.V1GcpAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsGcpGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsGcpGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsGcp() ([]*models.V1GcpAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsGcpListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsGcpList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1GcpAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1Client) GetCloudConfigGcp(configUID string) (*models.V1GcpCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsGcpGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsGcpGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) ImportClusterGcp(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersGcpImportParamsWithContext(h.Ctx).WithBody(
		&models.V1SpectroGcpClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1SpectroClustersGcpImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}
