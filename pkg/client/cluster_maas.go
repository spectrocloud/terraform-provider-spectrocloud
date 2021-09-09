package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)

func (h *V1alpha1Client) CreateClusterMaas(cluster *models.V1alpha1SpectroMaasClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersMaasCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersMaasCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

// TODO: no update for maas?
/*func (h *V1alpha1Client) UpdateClusterMaas(cluster *models.V1alpha1SpectroMaasClusterEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID
	params := clusterC.NewV1alpha1SpectroClustersMaasUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(cluster)
	_, err = client.V1alpha1SpectroClustersMaasUpdate(params)
	return err
}*/

func (h *V1alpha1Client) CreateMachinePoolMaas(cloudConfigId string, machinePool *models.V1alpha1MaasMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsMaasMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsMaasMachinePoolCreate(params)
	return err
}

// TODO: no update for maas?
/*func (h *V1alpha1Client) UpdateMachinePoolMaas(cloudConfigId string, machinePool *models.V1alpha1MassMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsMaasMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsMaasMachinePoolUpdate(params)
	return err
}*/

func (h *V1alpha1Client) DeleteMachinePoolMaas(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsMaasMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsMaasMachinePoolDelete(params)
	return err
}

// Cloud Account

func (h *V1alpha1Client) CreateCloudAccountMaas(account *models.V1alpha1MaasAccount) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1CloudAccountsMaasCreateParamsWithContext(h.ctx).WithBody(account)
	success, err := client.V1alpha1CloudAccountsMaasCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateCloudAccountMaas(account *models.V1alpha1MaasAccount) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1alpha1CloudAccountsMaasUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(account)
	_, err = client.V1alpha1CloudAccountsMaasUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteCloudAccountMaas(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudAccountsMaasDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1CloudAccountsMaasDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudAccountMaas(uid string) (*models.V1alpha1MaasAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsMaasGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1CloudAccountsMaasGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) GetCloudAccountsMaas() ([]*models.V1alpha1MaasAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsMaasListParamsWithContext(h.ctx)
	response, err := client.V1alpha1CloudAccountsMaasList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1alpha1MaasAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}

func (h *V1alpha1Client) GetCloudConfigMaas(configUID string) (*models.V1alpha1MaasCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsMaasGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsMaasGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

// TODO: no import?
/*func (h *V1alpha1Client) ImportClusterMaas(meta *models.V1ObjectMetaInputEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersMaasImportParamsWithContext(h.ctx).WithBody(
		&models.V1alpha1SpectroMaasClusterImportEntity{
			Metadata: meta,
		},
	)
	success, err := client.V1alpha1SpectroClustersMaasImport(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}*/
