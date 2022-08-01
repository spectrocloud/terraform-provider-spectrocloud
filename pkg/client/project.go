package client

import (
	"fmt"
	hapitransport "github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	hashboardC "github.com/spectrocloud/hapi/hashboard/client/v1"
	userC "github.com/spectrocloud/hapi/user/client/v1"
)

func (h *V1Client) CreateProject(body *models.V1ProjectEntity) (string, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return "", err
	}

	params := userC.NewV1ProjectsCreateParams().WithBody(body)
	success, err := client.V1ProjectsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) GetProjectUID(projectName string) (string, error) {
	projects, err := h.GetProjects()
	if err != nil {
		return "", err
	}

	for _, project := range projects.Items {
		if project.Metadata.Name == projectName {
			return project.Metadata.UID, nil
		}
	}

	return "", fmt.Errorf("project '%s' not found", projectName)
}

func (h *V1Client) GetProjectByUID(uid string) (*models.V1Project, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1ProjectsUIDGetParams().WithUID(uid)
	project, err := client.V1ProjectsUIDGet(params)
	if err != nil || project == nil {
		return nil, err
	}

	return project.Payload, nil
}

func (h *V1Client) GetProjects() (*models.V1ProjectsMetadata, error) {
	client, err := h.GetHashboard()
	if err != nil {
		return nil, err
	}

	params := hashboardC.NewV1ProjectsMetadataParams()

	projects, err := client.V1ProjectsMetadata(params)
	if err != nil || projects == nil {
		// to support 2.6 projects list
		if e, ok := err.(*hapitransport.TransportError); ok && e.HttpCode == 404 {
			limit := int64(0)
			userClient, err := h.GetUserClient()
			if err != nil {
				return nil, err
			}
			oldParams := userC.NewV1ProjectsListParams().WithLimit(&limit)
			oldProjects, err := userClient.V1ProjectsList(oldParams)
			if err != nil || oldProjects == nil {
				return nil, err
			}
			ret := make([]*models.V1ProjectMetadata, 0)
			for _, pr := range oldProjects.Payload.Items {
				ret = append(ret, &models.V1ProjectMetadata{
					Metadata: &models.V1ObjectEntity{
						UID:  pr.Metadata.UID,
						Name: pr.Metadata.Name,
					},
				})
			}
			return &models.V1ProjectsMetadata{
				Items: ret,
			}, nil
		}
		return nil, err
	}

	return projects.Payload, nil
}

func (h *V1Client) UpdateProject(uid string, body *models.V1ProjectEntity) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1ProjectsUIDUpdateParams().WithBody(body).WithUID(uid)
	_, err = client.V1ProjectsUIDUpdate(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1Client) DeleteProject(uid string) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1ProjectsUIDDeleteParams().WithUID(uid)
	_, err = client.V1ProjectsUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}
