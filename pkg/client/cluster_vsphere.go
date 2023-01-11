package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterVsphere(cluster *models.V1SpectroVsphereClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersVsphereCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersVsphereCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolVsphere(cloudConfigId string, machinePool *models.V1VsphereMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsVsphereMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsVsphereMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolVsphere(cloudConfigId string, machinePool *models.V1VsphereMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsVsphereMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsVsphereMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolVsphere(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsVsphereMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsVsphereMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1Client) CreateCloudAccountVsphere(account *models.V1VsphereAccount) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1CloudAccountsVsphereCreateParamsWithContext(h.Ctx).WithBody(account)
	success, err := client.V1CloudAccountsVsphereCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountVsphere(account *models.V1VsphereAccount) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1CloudAccountsVsphereUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
	_, err = client.V1CloudAccountsVsphereUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountVsphere(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudAccountsVsphereDeleteParamsWithContext(h.Ctx).WithUID(uid)
	_, err = client.V1CloudAccountsVsphereDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountVsphere(uid string) (*models.V1VsphereAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	getParams := clusterC.NewV1CloudAccountsVsphereGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsVsphereGet(getParams)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsVsphere() ([]*models.V1VsphereAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsVsphereListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsVsphereList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1VsphereAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1Client) GetCloudConfigVsphere(configUID string) (*models.V1VsphereCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsVsphereGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsVsphereGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) ImportClusterVsphere(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersVsphereImportParamsWithContext(h.Ctx).WithBody(
		&models.V1SpectroVsphereClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1SpectroClustersVsphereImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) ImportClusterGeneric(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersGenericImportParamsWithContext(h.Ctx).WithBody(
		&models.V1SpectroGenericClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1SpectroClustersGenericImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) GetVsphereClouldConfigValues(uid string) (*models.V1VsphereCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsVsphereGetParamsWithContext(h.Ctx).WithConfigUID(uid)
	cloudConfig, err := client.V1CloudConfigsVsphereGet(params)

	return cloudConfig.Payload, nil
}

func (h *V1Client) UpdateVsphereCloudConfigValues(uid string, clusterConfig *models.V1VsphereCloudClusterConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsVsphereUIDClusterConfigParamsWithContext(h.Ctx).WithConfigUID(uid).WithBody(clusterConfig)
	_, err = client.V1CloudConfigsVsphereUIDClusterConfig(params)

	return err
}
