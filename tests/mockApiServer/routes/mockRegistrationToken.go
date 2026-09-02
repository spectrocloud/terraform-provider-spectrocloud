package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// Fixture values the positive registration-token routes echo back.
// Kept as constants so tests can assert on them without duplicating the
// string literals. The expiry is fixed rather than "now + N days" so
// TestMain doesn't need Date.now() reproducibility.
const (
	mockRegTokenUID    = "test-reg-token-uid"
	mockRegTokenName   = "test-reg-token"
	mockRegTokenValue  = "eyJfake.reg.token"
	mockRegTokenExpiry = "2030-01-01"
)

func mockRegTokenPayload() *models.V1EdgeToken {
	expiry, _ := time.Parse("2006-01-02", mockRegTokenExpiry)
	return &models.V1EdgeToken{
		Metadata: &models.V1ObjectMeta{
			UID:  mockRegTokenUID,
			Name: mockRegTokenName,
			Annotations: map[string]string{
				"description": "mock registration token",
			},
		},
		Spec: &models.V1EdgeTokenSpec{
			Expiry: models.V1Time(strfmt.DateTime(expiry)),
			Token:  mockRegTokenValue,
			DefaultProject: &models.V1EdgeTokenProject{
				Name: "Default",
				UID:  "test-project-uid",
			},
		},
		Status: &models.V1EdgeTokenStatus{IsActive: true},
	}
}

func mockRegTokenList() *models.V1EdgeTokens {
	return &models.V1EdgeTokens{Items: []*models.V1EdgeToken{mockRegTokenPayload()}}
}

// RegistrationTokenRoutes wires POST/GET/PUT/DELETE for edge registration
// tokens. Note: resourceRegistrationTokenCreate issues both a POST (create)
// AND a follow-up GET on the list endpoint (via c.GetRegistrationToken →
// V1EdgeTokensList) to resolve the freshly-created token's value —
// hence the /v1/edgehosts/tokens GET route below is required for Create
// to complete, not just for standalone Reads.
func RegistrationTokenRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/edgehosts/tokens",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr(mockRegTokenUID)},
			},
		},
		{
			// List — hit by CreateRegistrationToken and by
			// GetRegistrationTokenByName during import.
			Method: "GET",
			Path:   "/v1/edgehosts/tokens",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockRegTokenList(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/edgehosts/tokens/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockRegTokenPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/edgehosts/tokens/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/edgehosts/tokens/{uid}/state",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/edgehosts/tokens/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

// RegistrationTokenNegativeRoutes returns error responses for every
// registration-token endpoint. The GET-by-UID variant uses the sentinel
// code "ResourceNotFound" so handleReadError takes the clear-ID branch;
// other endpoints use numeric codes so the diagnostic surfaces to the
// caller (matching how the API actually behaves for validation errors).
func RegistrationTokenNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/edgehosts/tokens",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "Registration token already exists"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/edgehosts/tokens",
			Response: ResponseData{
				StatusCode: http.StatusInternalServerError,
				Payload:    getError(strconv.Itoa(http.StatusInternalServerError), "Failed to list registration tokens"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/edgehosts/tokens/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError("ResourceNotFound", "Registration token not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/edgehosts/tokens/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid registration token update"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/edgehosts/tokens/{uid}/state",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid state transition"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/edgehosts/tokens/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Registration token not found"),
			},
		},
	}
}
