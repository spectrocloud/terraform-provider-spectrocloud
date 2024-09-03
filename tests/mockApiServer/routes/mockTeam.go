package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
)

func TeamRoutes() []Route {
	return []Route{
		{
			Method: "DELETE",
			Path:   "/v1/teams/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/teams/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/teams/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1Team{
					Metadata: &models.V1ObjectMeta{
						Name: "team-name",
						UID:  "team-123",
					},
					Spec: &models.V1TeamSpec{
						Roles:   []string{"role1"},
						Sources: []string{"source1"},
						Users:   []string{"user1"},
					},
					Status: nil,
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/teams",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]interface{}{"UID": "team-123"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/teams/{uid}/projects",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ProjectRolesEntity{
					Projects: []*models.V1UIDRoleSummary{
						{
							InheritedRoles: nil,
							Name:           "testadmin",
							Roles: []*models.V1UIDSummary{
								{
									Name: "test-role",
									UID:  "test-role-123",
								},
							},
							UID: "test-role-sum-id",
						},
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/teams/{uid}/projects",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/teams/{uid}/roles",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1TeamTenantRolesEntity{
					Roles: []*models.V1UIDSummary{
						{
							Name: "test-tenant-name",
							UID:  "test-tenant-id",
						},
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/teams/{uid}/roles",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/workspaces/teams/{teamUid}/roles",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1WorkspaceScopeRoles{
					Projects: []*models.V1ProjectsWorkspaces{
						{
							Name: "test-pjt-wp",
							UID:  "test-id1",
							Workspaces: []*models.V1WorkspacesRoles{
								{
									InheritedRoles: nil,
									Name:           "test-ws-name",
									Roles: []*models.V1WorkspaceRolesUIDSummary{
										{
											Name: "test-es-role-name",
											UID:  "test-id2",
										},
									},
									UID: "test-id3",
								},
							},
						},
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/workspaces/teams/{teamUid}/roles",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
	}
}
