package client

import (
	"github.com/spectrocloud/hapi/models"

	userC "github.com/spectrocloud/hapi/user/client/v1alpha1"
)

func (h *V1alpha1Client) CreateProject(body *models.V1alpha1ProjectEntity) (string, error) {
	client, err := h.getUserClient()
	if err != nil {
		return "", err
	}

	params := userC.NewV1alpha1ProjectsCreateParams().WithBody(body)
	success, err := client.V1alpha1ProjectsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) GetProject(uid string) (*models.V1alpha1Project, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	limit := int64(5000)
	params := userC.NewV1alpha1ProjectsListParams().WithLimit(&limit)
	projects, err := client.V1alpha1ProjectsList(params)
	if err != nil {
		return nil, err
	}

	for _, project := range projects.Payload.Items {
		if project.Metadata.UID == uid {
			return project, nil
		}
	}

	return nil, nil
}

func (h *V1alpha1Client) UpdateProject(uid string, body *models.V1alpha1ProjectEntity) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1ProjectsUIDUpdateParams().WithBody(body).WithUID(uid)
	_, err = client.V1alpha1ProjectsUIDUpdate(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1alpha1Client) DeleteProject(uid string) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1ProjectsUIDDeleteParams().WithUID(uid)
	_, err = client.V1alpha1ProjectsUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}
