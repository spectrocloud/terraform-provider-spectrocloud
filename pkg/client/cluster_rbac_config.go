package client

import (
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) GetClusterRbacConfig(uid string) (*models.V1ClusterRbacs, error) {
	if h.GetClusterRbacConfigFn != nil {
		return h.GetClusterRbacConfigFn(uid)
	}
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1SpectroClustersUIDConfigRbacsGetParamsWithContext(h.Ctx).WithUID(uid)
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
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1WorkspacesClusterRbacCreateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1WorkspacesClusterRbacCreate(params)
	return err
}

func (h *V1Client) UpdateClusterRbacConfig(uid string, config *models.V1ClusterRbacResourcesUpdateEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1SpectroClustersUIDConfigRbacsUpdateParamsWithContext(h.Ctx).WithUID(uid).WithBody(config)
	_, err = client.V1SpectroClustersUIDConfigRbacsUpdate(params)
	return err
}

func (h *V1Client) ApplyClusterRbacConfig(uid string, config []*models.V1ClusterRbacInputEntity) error {
	if rbac, err := h.GetClusterRbacConfig(uid); err != nil {
		return err
	} else if rbac == nil {
		return h.CreateClusterRbacConfig(uid, toCreateClusterRbac(config))
	} else {
		return h.UpdateClusterRbacConfig(uid, &models.V1ClusterRbacResourcesUpdateEntity{
			Rbacs: config,
		})
	}
}

func toCreateClusterRbac(rbacs []*models.V1ClusterRbacInputEntity) *models.V1ClusterRbac {
	bindings := make([]*models.V1ClusterRbacBinding, 0)

	for _, rbac := range rbacs {
		for _, binding := range rbac.Spec.Bindings {
			bindings = append(bindings, binding)
		}
	}

	return &models.V1ClusterRbac{
		Spec: &models.V1ClusterRbacSpec{
			Bindings: bindings,
		},
	}
}
