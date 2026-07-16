package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// mockPasswordPolicyPayload matches what resourcePasswordPolicyRead
// expects after Create: values echoed back so flattenPasswordPolicy
// reproduces the state the test set with prepare.
func mockPasswordPolicyPayload() *models.V1TenantPasswordPolicyEntity {
	return &models.V1TenantPasswordPolicyEntity{
		IsRegex:                   false,
		Regex:                     "",
		ExpiryDurationInDays:      90,
		FirstReminderInDays:       10,
		MinLength:                 8,
		MinNumOfBlockLetters:      1,
		MinNumOfDigits:            1,
		MinNumOfSmallLetters:      1,
		MinNumOfSpecialCharacters: 1,
	}
}

// PasswordPolicyRoutes wires the tenant-scoped password-policy endpoints.
// Note that per the API contract, Update is served over POST (not PUT) —
// there's no dedicated create endpoint, the resource treats Create as
// "upsert via Update", so all three of Create/Update/Delete on the
// Terraform side end up issuing the same POST here.
func PasswordPolicyRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/password/policy",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/password/policy",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockPasswordPolicyPayload(),
			},
		},
	}
}

func PasswordPolicyNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/password/policy",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid password policy"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/password/policy",
			Response: ResponseData{
				StatusCode: http.StatusInternalServerError,
				Payload:    getError(strconv.Itoa(http.StatusInternalServerError), "Failed to read password policy"),
			},
		},
	}
}
