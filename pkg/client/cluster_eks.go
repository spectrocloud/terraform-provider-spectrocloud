package client

import (
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)

func (h *V1alpha1Client) CreateClusterEks(cluster *models.V1alpha1SpectroEksClusterEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1alpha1SpectroClustersEksCreateParamsWithContext(h.ctx).WithBody(cluster)
	success, err := client.V1alpha1SpectroClustersEksCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateClusterEks(cluster *models.V1alpha1SpectroEksClusterEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	uid := cluster.Metadata.UID
	params := clusterC.NewV1alpha1SpectroClustersEksUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(cluster)
	_, err = client.V1alpha1SpectroClustersEksUpdate(params)
	return err
}


func (h *V1alpha1Client) CreateMachinePoolEks(cloudConfigId string, machinePool *models.V1alpha1EksMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsEksMachinePoolCreateParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsEksMachinePoolCreate(params)
	return err
}

func (h *V1alpha1Client) UpdateMachinePoolEks(cloudConfigId string, machinePool *models.V1alpha1EksMachinePoolConfigEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsEksMachinePoolUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithMachinePoolName(*machinePool.PoolConfig.Name).
		WithBody(machinePool)
	_, err = client.V1alpha1CloudConfigsEksMachinePoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteMachinePoolEks(cloudConfigId string, machinePoolName string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1CloudConfigsEksMachinePoolDeleteParamsWithContext(h.ctx).WithConfigUID(cloudConfigId).WithMachinePoolName(machinePoolName)
	_, err = client.V1alpha1CloudConfigsEksMachinePoolDelete(params)
	return err
}

func (h *V1alpha1Client) UpdateFargateProfiles(cloudConfigId string, fargateProfiles *models.V1alpha1EksFargateProfiles) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}
	params := clusterC.NewV1alpha1CloudConfigsEksUIDFargateProfilesUpdateParamsWithContext(h.ctx).
		WithConfigUID(cloudConfigId).
		WithBody(fargateProfiles)
	_, err = client.V1alpha1CloudConfigsEksUIDFargateProfilesUpdate(params)
	return err
}

func (h *V1alpha1Client) GetCloudConfigEks(configUID string) (*models.V1alpha1EksCloudConfig, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1CloudConfigsEksGetParamsWithContext(h.ctx).WithConfigUID(configUID)
	success, err := client.V1alpha1CloudConfigsEksGet(params)
	if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
		// TODO(saamalik) check with team if this is proper?
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
