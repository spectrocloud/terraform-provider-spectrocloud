package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)

func (h *V1alpha1Client) CreateClusterAws(cluster *models.V1alpha1SpectroAwsClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersAwsCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersAwsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateClusterAws(cluster *models.V1alpha1SpectroAwsClusterEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID
	params := clusterC.NewV1alpha1SpectroClustersAwsUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(cluster)
	_, err = client.V1alpha1SpectroClustersAwsUpdate(params)
	return err
}

func (h *V1alpha1Client) CreateMachinePoolAws(cloudConfigId string, machinePool *models.V1alpha1AwsMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAwsMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsAwsMachinePoolCreate(params)
	return err
}

func (h *V1alpha1Client) UpdateMachinePoolAws(cloudConfigId string, machinePool *models.V1alpha1AwsMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAwsMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsAwsMachinePoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteMachinePoolAws(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsAwsMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsAwsMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1alpha1Client) CreateCloudAccountAws(account *models.V1alpha1AwsAccount) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1CloudAccountsAwsCreateParamsWithContext(h.ctx).WithBody(account)
	success, err := client.V1alpha1CloudAccountsAwsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateCloudAccountAws(account *models.V1alpha1AwsAccount) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1alpha1CloudAccountsAwsUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(account)
	_, err = client.V1alpha1CloudAccountsAwsUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteCloudAccountAws(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudAccountsAwsDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1CloudAccountsAwsDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudAccountAws(uid string) (*models.V1alpha1AwsAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsAwsGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1CloudAccountsAwsGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) GetCloudAccountsAws() ([]*models.V1alpha1AwsAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsAwsListParamsWithContext(h.ctx)
	response, err := client.V1alpha1CloudAccountsAwsList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1alpha1AwsAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1alpha1Client) GetCloudConfigAws(configUID string) (*models.V1alpha1AwsCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsAwsGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsAwsGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) ImportClusterAws(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersAwsImportParamsWithContext(h.ctx).WithBody(
		&models.V1alpha1SpectroAwsClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1alpha1SpectroClustersAwsImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}
