package client

import (
	"github.com/spectrocloud/hapi/models"
	userC "github.com/spectrocloud/hapi/user/client/v1"
)

func (h *V1Client) CreateAlert(body *models.V1AlertEntity, projectUID string, component string) (string, error) {
	client, err := h.GetUserClient()

	if err != nil {
		return "", err
	}

	params := userC.NewV1ProjectsUIDAlertUpdateParams().WithBody(body).WithUID(projectUID).WithComponent(component)
	_, err = client.V1ProjectsUIDAlertUpdate(params)
	if err != nil {
		return "", err
	}
	return "nil", err

}

func (h *V1Client) UpdateAlert(body *models.V1AlertEntity, projectUID string, component string) (string, error) {
	client, err := h.GetUserClient()

	if err != nil {
		return "", err
	}

	params := userC.NewV1ProjectsUIDAlertUpdateParams().WithBody(body).WithUID(projectUID).WithComponent(component)
	_, err = client.V1ProjectsUIDAlertUpdate(params)
	if err != nil {
		return "", err
	}
	return "nill", err

}

func (h *V1Client) ReadAlert(body []*models.V1Channel, projectUID string, component string) ([]*models.V1Channel, error) {
	client, err := h.GetUserClient()
	channels := make([]*models.V1Channel, 0)
	if err != nil {
		return channels, err
	}

	for _, ch := range body {
		params := userC.NewV1ProjectsUIDAlertsUIDGetParams().WithUID(projectUID).WithComponent(component).WithAlertUID(ch.UID)
		resp, err := client.V1ProjectsUIDAlertsUIDGet(params)
		if err != nil {
			return channels, err
		}
		channels = append(channels, resp.Payload)
	}
	return channels, err

}

func (h *V1Client) DeleteAlerts(projectUID string, component string) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}
	params := userC.NewV1ProjectsUIDAlertDeleteParams().WithUID(projectUID).WithComponent(component)
	_, err = client.V1ProjectsUIDAlertDelete(params)
	if err != nil {
		return err
	}

	return nil
}
