package client

import (
	"github.com/spectrocloud/hapi/apiutil"
	"github.com/spectrocloud/hapi/models"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
)

func (h *V1Client) CreateIpPool(pcgUID string, pool *models.V1IPPoolInputEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", nil
	}

	params := clusterC.NewV1OverlordsUIDPoolCreateParams().WithUID(pcgUID).WithBody(pool)
	if resp, err := client.V1OverlordsUIDPoolCreate(params); err != nil {
		return "", err
	} else {
		return *resp.Payload.UID, nil
	}
}

func (h *V1Client) GetIpPool(pcgUID, poolUID string) (*models.V1IPPoolEntity, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1OverlordsUIDPoolsListParams().WithUID(pcgUID)
	if listResp, err := client.V1OverlordsUIDPoolsList(params); err != nil {
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

func (h *V1Client) UpdateIpPool(pcgUID, poolUID string, pool *models.V1IPPoolInputEntity) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1OverlordsUIDPoolUpdateParams().WithUID(pcgUID).
		WithBody(pool).WithPoolUID(poolUID)
	_, err = client.V1OverlordsUIDPoolUpdate(params)
	return err
}

func (h *V1Client) DeleteIpPool(pcgUID, poolUID string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1OverlordsUIDPoolDeleteParams().WithUID(pcgUID).WithPoolUID(poolUID)
	_, err = client.V1OverlordsUIDPoolDelete(params)
	return err
}
