package client

import (
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterRbacConfig(uid string) (*models.V1ClusterRbacs, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1SpectroClustersUIDConfigRbacsGetParamsWithContext(h.ctx).WithUID(uid)
	success, err := client.V1SpectroClustersUIDConfigRbacsGet(params)
	if err != nil {
		if herr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) CreateClusterRbacConfig(uid string, config *models.V1ClusterRbac) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1WorkspacesClusterRbacCreateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
	_, err = client.V1WorkspacesClusterRbacCreate(params)
	return err
}

func (h *V1Client) UpdateClusterRbacConfig(uid string, config *models.V1ClusterRbacResourcesUpdateEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1SpectroClustersUIDConfigRbacsUpdateParamsWithContext(h.ctx).WithUID(uid).WithBody(config)
	_, err = client.V1SpectroClustersUIDConfigRbacsUpdate(params)
	return err
}

func (h *V1Client) ApplyClusterRbacConfig(uid string, config *models.V1ClusterRbac) error {
	if rbac, err := h.GetClusterRbacConfig(uid); err != nil {
		return err
	} else if rbac == nil {
		return h.CreateClusterRbacConfig(uid, config)
	} else {
		return h.UpdateClusterRbacConfig(uid, toUpdateRbac(config))
	}
}

func toUpdateRbac(config *models.V1ClusterRbac) *models.V1ClusterRbacResourcesUpdateEntity {
	rbacs := make([]*models.V1ClusterRbacInputEntity, 0, 1)

	if config != nil {
		rbacs = append(rbacs, &models.V1ClusterRbacInputEntity{
			Spec: &models.V1ClusterRbacSpec{
				Bindings: config.Spec.Bindings,
			},
		})
	}

	return &models.V1ClusterRbacResourcesUpdateEntity{
		Rbacs: rbacs,
	}
}
