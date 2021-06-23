package client

import (
	"fmt"

	userC "github.com/spectrocloud/hapi/user/client/v1alpha1"

	"github.com/spectrocloud/hapi/models"
)

func (h *V1alpha1Client) GetRole(roleName string) (*models.V1alpha1Role, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1RolesListParams()
	roles, err := client.V1alpha1RolesList(params)
	if err != nil {
		return nil, err
	}

	for _, role := range roles.Payload.Items {
		if role.Metadata.Name == roleName {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role '%s' not found", roleName)
}

func (h *V1alpha1Client) GetUser(name string) (*models.V1alpha1User, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1UsersListParams()
	users, err := client.V1alpha1UsersList(params)
	if err != nil {
		return nil, err
	}

	for _, user := range users.Payload.Items {
		if user.Metadata.Name == name {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user '%s' not found", name)
}

func (h *V1alpha1Client) CreateTeam(team *models.V1alpha1Team) (string, error) {
	client, err := h.getUserClient()
	if err != nil {
		return "", err
	}

	params := userC.NewV1alpha1TeamsCreateParams().WithBody(team)
	success, err := client.V1alpha1TeamsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1alpha1Client) UpdateTeam(uid string, team *models.V1alpha1Team) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1TeamsUIDUpdateParams().WithBody(team).WithUID(uid)
	_, err = client.V1alpha1TeamsUIDUpdate(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1alpha1Client) DeleteTeam(uid string) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1TeamsUIDDeleteParams().WithUID(uid)
	_, err = client.V1alpha1TeamsUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1alpha1Client) AssociateTeamProjectRole(uid string, body *models.V1alpha1ProjectRolesPatch) error {
	client, err := h.getUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1alpha1TeamsProjectRolesPutParams().WithUID(uid).WithBody(body)
	_, err = client.V1alpha1TeamsProjectRolesPut(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1alpha1Client) GetTeamProjectRoleAssociation(uid string) (*models.V1alpha1ProjectRolesEntity, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1TeamsProjectRolesParams().WithUID(uid)
	success, err := client.V1alpha1TeamsProjectRoles(params)
	if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1alpha1Client) GetTeam(uid string) (*models.V1alpha1Team, error) {
	client, err := h.getUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1alpha1TeamsUIDGetParams().WithUID(uid)
	success, err := client.V1alpha1TeamsUIDGet(params)
	if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
