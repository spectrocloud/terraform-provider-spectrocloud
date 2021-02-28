package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)

func (h *V1alpha1Client) CreateClusterVsphere(cluster *models.V1alpha1SpectroVsphereClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersVsphereCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersVsphereCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateClusterVsphere(cluster *models.V1alpha1SpectroVsphereClusterEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID
	params := clusterC.NewV1alpha1SpectroClustersVsphereUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(cluster)
	_, err = client.V1alpha1SpectroClustersVsphereUpdate(params)
	return err
}

func (h *V1alpha1Client) CreateMachinePoolVsphere(cloudConfigId string, machinePool *models.V1alpha1VsphereMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsVsphereMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsVsphereMachinePoolCreate(params)
	return err
}

func (h *V1alpha1Client) UpdateMachinePoolVsphere(cloudConfigId string, machinePool *models.V1alpha1VsphereMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsVsphereMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsVsphereMachinePoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteMachinePoolVsphere(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsVsphereMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsVsphereMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1alpha1Client) CreateCloudAccountVsphere(account *models.V1alpha1VsphereAccount) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1CloudAccountsVsphereCreateParamsWithContext(h.ctx).WithBody(account)
	success, err := client.V1alpha1CloudAccountsVsphereCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateCloudAccountVsphere(account *models.V1alpha1VsphereAccount) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1alpha1CloudAccountsVsphereUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(account)
	_, err = client.V1alpha1CloudAccountsVsphereUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteCloudAccountVsphere(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudAccountsVsphereDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1CloudAccountsVsphereDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudAccountVsphere(uid string) (*models.V1alpha1VsphereAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	getParams := clusterC.NewV1alpha1CloudAccountsVsphereGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1CloudAccountsVsphereGet(getParams)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) GetCloudAccountsVsphere() ([]*models.V1alpha1VsphereAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsVsphereListParamsWithContext(h.ctx)
	response, err := client.V1alpha1CloudAccountsVsphereList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1alpha1VsphereAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1alpha1Client) GetCloudConfigVsphere(configUID string) (*models.V1alpha1VsphereCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsVsphereGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsVsphereGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) ImportClusterVsphere(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersVsphereImportParamsWithContext(h.ctx).WithBody(
		&models.V1alpha1SpectroVsphereClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1alpha1SpectroClustersVsphereImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}
