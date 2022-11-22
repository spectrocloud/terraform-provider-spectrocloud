package client

import (
	"github.com/spectrocloud/hapi/models"
	userC "github.com/spectrocloud/hapi/user/client/v1"
)

func (h *V1Client) CreateAlert(body *models.V1Channel, projectUID string, component string) (string, error) {
	client, err := h.GetUserClient()

	if err != nil {
		return "", err
	}

	params := userC.NewV1ProjectsUIDAlertCreateParams().WithBody(body).WithUID(projectUID).WithComponent(component)
	success, err := client.V1ProjectsUIDAlertCreate(params)
	if err != nil {
		return "", err
	}
	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateAlert(body *models.V1Channel, projectUID string, component string, alertUID string) (string, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return "", err
	}
	params := userC.NewV1ProjectsUIDAlertsUIDUpdateParams().WithBody(body).WithUID(projectUID).WithComponent(component).WithAlertUID(alertUID)
	_, err = client.V1ProjectsUIDAlertsUIDUpdate(params)
	if err != nil {
		return "", err
	}
	return "success", nil

}

func (h *V1Client) ReadAlert(projectUID string, component string, alertUID string) (*models.V1Channel, error) {
	client, err := h.GetUserClient()
	channel := &models.V1Channel{}
	if err != nil {
		return channel, err
	}
	params := userC.NewV1ProjectsUIDAlertsUIDGetParams().WithUID(projectUID).WithComponent(component).WithAlertUID(alertUID)
	success, err := client.V1ProjectsUIDAlertsUIDGet(params)
	if err != nil {
		return nil, err
	}
	return success.Payload, nil

}

func (h *V1Client) DeleteAlerts(projectUID string, component string, alertUID string) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}
	params := userC.NewV1ProjectsUIDAlertsUIDDeleteParams().WithUID(projectUID).WithComponent(component).WithAlertUID(alertUID)
	_, err = client.V1ProjectsUIDAlertsUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}
