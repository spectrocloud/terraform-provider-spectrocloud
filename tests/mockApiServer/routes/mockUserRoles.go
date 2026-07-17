package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// UserRolesRoutes fills in the endpoints needed by resource_user Read
// beyond what mockUsers.go already provides. resourceUserRead calls:
//   - GetUserSummaryByEmail   → POST /v1/users/summary
//   - GetUserProjectRole      → GET  /v1/users/{uid}/projects
//   - GetUserTenantRole       → GET  /v1/users/{uid}/roles
//   - GetUserWorkspaceRole    → GET  /v1/workspaces/users/{userUid}/roles
//   - GetUserResourceRoles    → GET  /v1/users/{uid}/resourceRoles
//
// Plus the write-side PUTs the Create/Update paths hit.
func UserRolesRoutes() []Route {
	summaryUID := "12345"
	summaryName := "test"
	return []Route{
		// GetUserSummaryByEmail
		{
			Method: "POST",
			Path:   "/v1/users/summary",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1UsersSummaryList{
					Items: []*models.V1UserSummary{
						{
							Metadata: &models.V1ObjectMeta{
								Name: summaryName,
								UID:  summaryUID,
							},
							Spec: &models.V1UserSpecSummary{
								FirstName: "test",
								LastName:  "spectro",
								EmailID:   "test@spectrocloud.com",
							},
						},
					},
				},
			},
		},

		// User project roles (GET and PUT)
		{
			Method: "GET",
			Path:   "/v1/users/{uid}/projects",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1ProjectRolesEntity{},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/{uid}/projects",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// User tenant roles (GET and PUT)
		{
			Method: "GET",
			Path:   "/v1/users/{uid}/roles",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1UserRolesEntity{},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/{uid}/roles",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// User workspace roles (GET and PUT)
		{
			Method: "GET",
			Path:   "/v1/workspaces/users/{userUid}/roles",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1WorkspaceScopeRoles{},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/workspaces/users/{userUid}/roles",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// User resource roles (GET, POST, DELETE)
		{
			Method: "GET",
			Path:   "/v1/users/{uid}/resourceRoles",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    &models.V1ResourceRolesUpdateEntity{},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/users/{uid}/resourceRoles",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/users/{uid}/resourceRoles/{resourceRoleUid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// User Create/Delete — mockUsers.go doesn't cover these.
		{
			Method: "POST",
			Path:   "/v1/users",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr("test-user-uid")},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/users/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}
