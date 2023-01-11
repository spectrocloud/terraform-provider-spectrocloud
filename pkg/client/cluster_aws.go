package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateClusterAws(cluster *models.V1SpectroAwsClusterEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersAwsCreateParamsWithContext(h.Ctx).WithBody(cluster)
	success, err := client.V1SpectroClustersAwsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) CreateMachinePoolAws(cloudConfigId string, machinePool *models.V1AwsMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsAwsMachinePoolCreateParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1CloudConfigsAwsMachinePoolCreate(params)
	return err
}

func (h *V1Client) UpdateMachinePoolAws(cloudConfigId string, machinePool *models.V1AwsMachinePoolConfigEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsAwsMachinePoolUpdateParamsWithContext(h.Ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1CloudConfigsAwsMachinePoolUpdate(params)
	return err
}

func (h *V1Client) DeleteMachinePoolAws(cloudConfigId string, machinePoolName string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1CloudConfigsAwsMachinePoolDeleteParamsWithContext(h.Ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1CloudConfigsAwsMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1Client) CreateCloudAccountAws(account *models.V1AwsAccount, ClusterContext string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	var params *clusterC.V1CloudAccountsAwsCreateParams
	switch ClusterContext {
	case "project":
		params = clusterC.NewV1CloudAccountsAwsCreateParamsWithContext(h.Ctx).WithBody(account)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsAwsCreateParams().WithBody(account)
		break
	default:
		break
	}
	success, err := client.V1CloudAccountsAwsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateCloudAccountAws(account *models.V1AwsAccount) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	uid := account.Metadata.UID
	var params *clusterC.V1CloudAccountsAwsUpdateParams
	switch account.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1CloudAccountsAwsUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(account)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsAwsUpdateParams().WithBody(account)
		break
	default:
		break
	}
	_, err = client.V1CloudAccountsAwsUpdate(params)
	return err
}

func (h *V1Client) DeleteCloudAccountAws(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	account, err := h.GetCloudAccountAws(uid)
	if err != nil {
		return err
	}

	var params *clusterC.V1CloudAccountsAwsDeleteParams
	switch account.Metadata.Annotations["scope"] {
	case "project":
		params = clusterC.NewV1CloudAccountsAwsDeleteParamsWithContext(h.Ctx).WithUID(uid)
		break
	case "tenant":
		params = clusterC.NewV1CloudAccountsAwsDeleteParams().WithUID(uid)
		break
	default:
		break
	}
	_, err = client.V1CloudAccountsAwsDelete(params)
	return err
}

func (h *V1Client) GetCloudAccountAws(uid string) (*models.V1AwsAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsAwsGetParamsWithContext(h.Ctx).WithUID(uid)
	success, err := client.V1CloudAccountsAwsGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetCloudAccountsAws() ([]*models.V1AwsAccount, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudAccountsAwsListParamsWithContext(h.Ctx)
	response, err := client.V1CloudAccountsAwsList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1AwsAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1Client) GetCloudConfigAws(configUID string) (*models.V1AwsCloudConfig, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1CloudConfigsAwsGetParamsWithContext(h.Ctx).WithConfigUID(configUID)
	success, err := client.V1CloudConfigsAwsGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) ImportClusterAws(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1SpectroClustersAwsImportParamsWithContext(h.Ctx).WithBody(
		&models.V1SpectroAwsClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1SpectroClustersAwsImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}
