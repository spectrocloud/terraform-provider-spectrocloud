package client

import (
	"fmt"

	userC "github.com/spectrocloud/hapi/user/client/v1"

	"github.com/spectrocloud/hapi/models"
)

func (h *V1Client) GetRole(roleName string) (*models.V1Role, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1RolesListParams()
	roles, err := client.V1RolesList(params)
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

func (h *V1Client) GetUser(name string) (*models.V1User, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1UsersListParams()
	users, err := client.V1UsersList(params)
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

func (h *V1Client) CreateTeam(team *models.V1Team) (string, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return "", err
	}

	params := userC.NewV1TeamsCreateParams().WithBody(team)
	success, err := client.V1TeamsCreate(params)
	if err != nil {
		return "", err
	}

	return *success.Payload.UID, nil
}

func (h *V1Client) UpdateTeam(uid string, team *models.V1Team) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1TeamsUIDUpdateParams().WithBody(team).WithUID(uid)
	_, err = client.V1TeamsUIDUpdate(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1Client) DeleteTeam(uid string) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1TeamsUIDDeleteParams().WithUID(uid)
	_, err = client.V1TeamsUIDDelete(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1Client) AssociateTeamProjectRole(uid string, body *models.V1ProjectRolesPatch) error {
	client, err := h.GetUserClient()
	if err != nil {
		return err
	}

	params := userC.NewV1TeamsProjectRolesPutParams().WithUID(uid).WithBody(body)
	_, err = client.V1TeamsProjectRolesPut(params)
	if err != nil {
		return err
	}

	return nil
}

func (h *V1Client) GetTeamProjectRoleAssociation(uid string) (*models.V1ProjectRolesEntity, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1TeamsProjectRolesParams().WithUID(uid)
	success, err := client.V1TeamsProjectRoles(params)
	if err != nil {
		return nil, err
	}

	return success.Payload, nil
}

func (h *V1Client) GetTeam(uid string) (*models.V1Team, error) {
	client, err := h.GetUserClient()
	if err != nil {
		return nil, err
	}

	params := userC.NewV1TeamsUIDGetParams().WithUID(uid)
	success, err := client.V1TeamsUIDGet(params)
	if err != nil {
		return nil, err
	}

	return success.Payload, nil
}
