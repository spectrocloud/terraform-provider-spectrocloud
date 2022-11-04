package client

import (
	"fmt"
	hashboardC "github.com/spectrocloud/hapi/hashboard/client/v1"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client/herr"
)

func (h *V1Client) GetApplicationProfileByName(profileName string) (*models.V1AppProfileSummary, error) {
	client, err := h.GetHashboard()
	if err != nil {
		return nil, err
	}

	limit := int64(0)
	params := hashboardC.NewV1DashboardAppProfilesParamsWithContext(h.Ctx).WithLimit(&limit)
	profiles, err := client.V1DashboardAppProfiles(params)
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles.Payload.AppProfiles {
		if profile.Metadata.Name == profileName {
			return profile, nil
		}
	}

	return nil, fmt.Errorf("Application profile '%s' not found.", profileName)
}

func (h *V1Client) GetApplicationProfile(uid string) (*models.V1EdgeHostDevice, error) {
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

func (h *V1Client) CreateApplicationProfile(createHostDevice *models.V1EdgeHostDeviceEntity) (string, error) {
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

func (h *V1Client) UpdateApplicationProfile(uid string, appliance *models.V1EdgeHostDevice) error {
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

func (h *V1Client) DeleteApplicationProfile(uid string) error {
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
