package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"
)

func (h *V1alpha1Client) CreateClusterOpenStack(cluster *models.V1alpha1SpectroOpenStackClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersOpenStackCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersOpenStackCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) CreateCloudAccountOpenStack(account *models.V1alpha1OpenStackAccount) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1CloudAccountsOpenStackCreateParamsWithContext(h.ctx).WithBody(account)
	success, err := client.V1alpha1CloudAccountsOpenStackCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) CreateMachinePoolOpenStack(cloudConfigId string, machinePool *models.V1alpha1OpenStackMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsOpenStackMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsOpenStackMachinePoolCreate(params)
	return err
}

func (h *V1alpha1Client) UpdateMachinePoolOpenStack(cloudConfigId string, machinePool *models.V1alpha1OpenStackMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsOpenStackMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsOpenStackMachinePoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteMachinePoolOpenStack(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsOpenStackMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsOpenStackMachinePoolDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudAccountOpenStack(uid string) (*models.V1alpha1OpenStackAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsOpenStackGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1alpha1CloudAccountsOpenStackGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) GetCloudConfigOpenStack(configUID string) (*models.V1alpha1OpenStackCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsOpenStackGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsOpenStackGet(params)

	if herr.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}


func (h *V1alpha1Client) UpdateCloudAccountOpenStack(account *models.V1alpha1OpenStackAccount) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := account.Metadata.UID
	params := clusterC.NewV1alpha1CloudAccountsOpenStackUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(account)
	_, err = client.V1alpha1CloudAccountsOpenStackUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteCloudAccountOpenStack(uid string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudAccountsOpenStackDeleteParamsWithContext(h.ctx).WithUID(uid)
	_, err = client.V1alpha1CloudAccountsOpenStackDelete(params)
	return err
}

func (h *V1alpha1Client) GetCloudAccountsOpenStack() ([]*models.V1alpha1OpenStackAccount, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudAccountsOpenStackListParamsWithContext(h.ctx)
	response, err := client.V1alpha1CloudAccountsOpenStackList(params)
	if err != nil {
		return nil, err
	}

	accounts := make([]*models.V1alpha1OpenStackAccount, len(response.Payload.Items))
	for i, account := range response.Payload.Items {
		accounts[i] = account
	}

	return accounts, nil
}


