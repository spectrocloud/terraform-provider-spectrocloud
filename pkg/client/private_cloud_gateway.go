package client

import (
	"errors"
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

// get ip pool by uid
func (h *V1Client) GetIpPool(pcgUID, poolUID string) (*models.V1IPPoolEntity, error) {
	pools, err := h.GetIpPools(pcgUID)
	if err != nil {
		return nil, err
	}
	for _, pool := range pools {
		if pool.Metadata.UID == poolUID {
			return pool, nil
		}
	}
	return nil, nil
}

// get ip pool by name
func (h *V1Client) GetIpPoolByName(pcgUID, poolName string) (*models.V1IPPoolEntity, error) {
	pools, err := h.GetIpPools(pcgUID)
	if err != nil {
		return nil, err
	}
	for _, pool := range pools {
		if pool.Metadata.Name == poolName {
			return pool, nil
		}
	}
	return nil, errors.New("ip pool not found: " + poolName)
}

func (h *V1Client) GetIpPools(pcgUID string) ([]*models.V1IPPoolEntity, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1OverlordsUIDPoolsListParams().WithUID(pcgUID)
	listResp, err := client.V1OverlordsUIDPoolsList(params)
	if err != nil {
		if v1Err := apiutil.ToV1ErrorObj(err); v1Err.Code != "ResourceNotFound" {
			return nil, err
		}
	}
	return listResp.Payload.Items, nil
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

func (h *V1Client) GetPrivateCloudGatewayID(name *string) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1OverlordsListParams()
	listResp, err := client.V1OverlordsList(params)
	if err != nil {
		return "", err
	}

	for _, pcg := range listResp.Payload.Items {
		if pcg.Metadata.Name == *name {
			return pcg.Metadata.UID, nil
		}
	}
	// return not found error
	return "", errors.New("Private Cloud Gateway not found: " + *name)
}
