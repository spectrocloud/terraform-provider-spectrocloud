package client

import (
	"github.com/spectrocloud/hapi/apiutil"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1alpha1"
)

func (h *V1alpha1Client) CreateIpPool(pcgUID string, pool *models.V1alpha1IPPoolInputEntity) (string, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return "", nil
	}

	params := clusterC.NewV1alpha1OverlordsUIDPoolCreateParams().WithUID(pcgUID).WithBody(pool)
	if resp, err := client.V1alpha1OverlordsUIDPoolCreate(params); err != nil {
		return "", err
	} else {
		return *resp.Payload.UID, nil
	}
}

func (h *V1alpha1Client) GetIpPool(pcgUID, poolUID string) (*models.V1alpha1IPPoolEntity, error) {
	client, err := h.getClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1alpha1OverlordsUIDPoolsListParams().WithUID(pcgUID)
	if listResp, err := client.V1alpha1OverlordsUIDPoolsList(params); err != nil {
		if v1Err := apiutil.ToV1ErrorObj(err); v1Err.Code != "ResourceNotFound" {
			return nil, err
		}
	} else if listResp.Payload != nil && listResp.Payload.Items != nil {
		for _, pool := range listResp.Payload.Items {
			if p := *pool; pool.Metadata.UID == poolUID {
				return &p, nil
			}
		}
	}
	return nil, nil
}

func (h *V1alpha1Client) UpdateIpPool(pcgUID, poolUID string, pool *models.V1alpha1IPPoolInputEntity) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1OverlordsUIDPoolUpdateParams().WithUID(pcgUID).
		WithBody(pool).WithPoolUID(poolUID)
	_, err = client.V1alpha1OverlordsUIDPoolUpdate(params)
	return err
}

func (h *V1alpha1Client) DeleteIpPool(pcgUID, poolUID string) error {
	client, err := h.getClusterClient()
	if err != nil {
		return nil
	}

	params := clusterC.NewV1alpha1OverlordsUIDPoolDeleteParams().WithUID(pcgUID).WithPoolUID(poolUID)
	_, err = client.V1alpha1OverlordsUIDPoolDelete(params)
	return err
}
