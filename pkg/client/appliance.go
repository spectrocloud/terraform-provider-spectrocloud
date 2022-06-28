package client

import (
	"fmt"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"
)

func (h *V1Client) GetApplianceByName(deviceName string) (*models.V1EdgeHostDevice, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	//limit := int64(0)
	//params := clusterC.NewV1EdgeHostDevicesListParamsWithContext(h.ctx).WithLimit(&limit)
	params := clusterC.NewV1EdgeHostDevicesListParamsWithContext(h.Ctx)
	appliances, err := client.V1EdgeHostDevicesList(params)
	if err != nil {
		return nil, err
	}

	for _, appliance := range appliances.Payload.Items {
		if appliance.Metadata.Name == deviceName {
			return appliance, nil
		}
	}

	return nil, fmt.Errorf("Appliance '%s' not found.", deviceName)
}

func (h *V1Client) GetAppliance(uid string) (*models.V1EdgeHostDevice, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return nil, err
	}

	params := clusterC.NewV1EdgeHostDevicesUIDGetParamsWithContext(h.Ctx).WithUID(uid)
	response, err := client.V1EdgeHostDevicesUIDGet(params)
	if err != nil {
		if herr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return response.Payload, nil
}

func (h *V1Client) CreateAppliance(createHostDevice *models.V1EdgeHostDeviceEntity) (string, error) {
	client, err := h.GetClusterClient()
	if err != nil {
		return "", err
	}

	params := clusterC.NewV1EdgeHostDevicesCreateParams().WithContext(h.Ctx).WithBody(createHostDevice)
	if resp, err := client.V1EdgeHostDevicesCreate(params); err != nil {
		return "", err
	} else {
		return *resp.Payload.UID, nil
	}
}

func (h *V1Client) UpdateAppliance(uid string, appliance *models.V1EdgeHostDevice) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1EdgeHostDevicesUIDUpdateParams().WithContext(h.Ctx).WithBody(appliance).WithUID(uid)
	_, err = client.V1EdgeHostDevicesUIDUpdate(params)
	if err != nil && !herr.IsEdgeHostDeviceNotRegistered(err) {
		return err
	}

	return nil
}

func (h *V1Client) DeleteAppliance(uid string) error {
	client, err := h.GetClusterClient()
	if err != nil {
		return err
	}

	params := clusterC.NewV1EdgeHostDevicesUIDDeleteParams().WithContext(h.Ctx).WithUID(uid)
	_, err = client.V1EdgeHostDevicesUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}
