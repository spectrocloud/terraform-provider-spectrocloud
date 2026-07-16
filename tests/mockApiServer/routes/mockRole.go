package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// mockRoleUID is the fixed UID Create returns and GetRoleByID matches.
// Kept stable (rather than generateRandomStringUID()) so tests can assert
// on d.Id() after Create without threading the response value through.
const mockRoleUID = "test-role-uid"

func mockRolePayload() *models.V1Role {
	return &models.V1Role{
		Metadata: &models.V1ObjectMeta{
			Name: "test-role",
			UID:  mockRoleUID,
			Annotations: map[string]string{
				"scope": "project",
			},
		},
		Spec: &models.V1RoleSpec{
			Permissions: []string{"perm1", "perm2"},
			Scope:       "project",
			Type:        "user",
		},
		Status: &models.V1RoleStatus{IsEnabled: true},
	}
}

func getRolesList() *models.V1Roles {
	return &models.V1Roles{
		Items: []*models.V1Role{mockRolePayload()},
		Listmeta: &models.V1ListMetaData{
			Continue: "",
			Count:    1,
			Limit:    0,
			Offset:   0,
		},
	}
}

func RolesRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/roles",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr(mockRoleUID)},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/roles",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getRolesList(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/roles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockRolePayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/roles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/roles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

func RolesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/roles",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "Role already exists"),
			},
		},
		{
			// List gets its own 404 — the existing pre-Wave-3 behavior we
			// preserved so any downstream tests that were relying on it
			// don't break.
			Method: "GET",
			Path:   "/v1/roles",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "No roles are found"),
			},
		},
		{
			// GET-by-UID uses the ResourceNotFound sentinel so
			// handleReadError takes its clear-ID branch.
			Method: "GET",
			Path:   "/v1/roles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError("ResourceNotFound", "Role not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/roles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid role update"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/roles/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Role not found"),
			},
		},
	}
}
