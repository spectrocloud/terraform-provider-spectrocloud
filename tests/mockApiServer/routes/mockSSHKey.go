package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// mockSSHKeyName / mockSSHKeyPublicKey are the values the positive Read
// route reports back. Kept as package-level constants so a test can assert
// on them without duplicating literals.
const (
	mockSSHKeyName      = "test-ssh-key"
	mockSSHKeyPublicKey = "ssh-rsa AAAAB3NzaC1yc2ETEST test-ssh-key"
)

func mockSSHKeyPayload() *models.V1UserAssetSSH {
	return &models.V1UserAssetSSH{
		Metadata: &models.V1ObjectMeta{
			Name: mockSSHKeyName,
			UID:  "test-ssh-key-uid",
		},
		Spec: &models.V1UserAssetSSHSpec{
			PublicKey: mockSSHKeyPublicKey,
		},
	}
}

// SSHKeyRoutes returns the happy-path route set for the SSH key
// resource. Create returns a fixed UID so tests can predict d.Id() after
// Create; Read echoes back the same fixture so a subsequent flatten
// matches the state we set on create.
func SSHKeyRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/users/assets/sshkeys",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr("test-ssh-key-uid")},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/sshkeys/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockSSHKeyPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/sshkeys/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/users/assets/sshkeys/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

// SSHKeyNegativeRoutes returns error responses for every SSH endpoint so
// TestResourceSSHKeyCRUDNegative can assert that each CRUD op propagates
// the server-side error through resource*'s diag.Diagnostics.
func SSHKeyNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/users/assets/sshkeys",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "SSH key already exists"),
			},
		},
		{
			// GET → 404 with the sentinel Code "ResourceNotFound" (the string
			// herr.IsNotFound matches on) so handleReadError takes its
			// clear-ID branch instead of surfacing the error to Terraform.
			// This is deliberately distinct from Delete's 404 below, which
			// uses a numeric code so its diag survives.
			Method: "GET",
			Path:   "/v1/users/assets/sshkeys/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError("ResourceNotFound", "SSH key not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/sshkeys/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid SSH key update"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/users/assets/sshkeys/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "SSH key not found"),
			},
		},
	}
}
