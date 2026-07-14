package routes

import (
	"net/http"
	"strconv"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// Fixture UIDs. Kept stable so tests can assert on d.Id() after Create.
const (
	mockCloudWatchAuditTrailUID = "test-cw-audit-uid"
	mockSplunkAuditTrailUID     = "test-splunk-audit-uid"
)

func mockCloudWatchDataSinkPayload() *models.V1DataSinkConfig {
	partition := "aws"
	credType := models.V1AwsCloudAccountCredentialTypeSecret
	return &models.V1DataSinkConfig{
		Metadata: &models.V1ObjectMeta{
			Name: "test-cw-audit",
			UID:  mockCloudWatchAuditTrailUID,
		},
		Spec: &models.V1DataSinkSpec{
			AuditDataSinks: []*models.V1DataSinkableSpec{
				{
					Type: models.V1DataSinkableSpecTypeCloudwatch,
					CloudWatch: &models.V1CloudWatch{
						Group:  "logs",
						Region: "us-east-1",
						Stream: "audit",
						Credentials: &models.V1AwsCloudAccount{
							AccessKey:      "AKIATEST",
							SecretKey:      "**** (server-side sanitized)",
							Partition:      &partition,
							CredentialType: &credType,
						},
					},
				},
			},
		},
	}
}

func mockSplunkSinkPayload() *models.V1SplunkSink {
	hec := "https://splunk.example.com:8088"
	return &models.V1SplunkSink{
		Metadata: &models.V1ObjectMeta{
			Name: "test-splunk-audit",
			UID:  mockSplunkAuditTrailUID,
		},
		Spec: &models.V1SplunkSinkSpec{
			HecURL: &hec,
			Index:  "main",
			Source: "palette-audit",
		},
	}
}

// AuditTrailRoutes wires both CloudWatch and Splunk audit-trail endpoints,
// plus the pre-create validate calls each side issues. Note: Create for
// CloudWatch does POST /assets/dataSinks, but per the SDK it first calls
// /cloudwatch/validate. Both must succeed for the Terraform resource's
// Create/Update to advance.
func AuditTrailRoutes() []Route {
	// Constant response used by the validate endpoints. They return
	// 204 No Content on success, per the API contract.
	validateOK := ResponseData{StatusCode: http.StatusNoContent, Payload: nil}

	return []Route{
		// CloudWatch validate — POST /v1/clouds/aws/cloudwatch/validate
		{Method: "POST", Path: "/v1/clouds/aws/cloudwatch/validate", Response: validateOK},

		// CloudWatch CRUD — data sinks
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/assets/dataSinks",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr(mockCloudWatchAuditTrailUID)},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/assets/dataSinks",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockCloudWatchDataSinkPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/assets/dataSinks",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/tenants/{tenantUid}/assets/dataSinks",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},

		// Splunk validate — POST /v1/tenants/{tenantUid}/datasinks/splunk/validate
		{Method: "POST", Path: "/v1/tenants/{tenantUid}/datasinks/splunk/validate", Response: validateOK},

		// Splunk CRUD
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr(mockSplunkAuditTrailUID)},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockSplunkSinkPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

// AuditTrailNegativeRoutes surfaces validate-failure errors on both sides.
// Read paths remain "OK-shaped" on the negative server for the same reason
// as password_policy/developer_setting — GetCloudWatchAuditTrail uses
// apiutil.Is404() to handle NotFound but any other status can still
// panic through nested nils in the response chain; keeping GET valid
// avoids that noise.
func AuditTrailNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/clouds/aws/cloudwatch/validate",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid CloudWatch credentials"),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk/validate",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid Splunk configuration"),
			},
		},
		{
			// DELETE is reachable without validate — cover it directly.
			Method: "DELETE",
			Path:   "/v1/tenants/{tenantUid}/assets/dataSinks",
			Response: ResponseData{
				StatusCode: http.StatusInternalServerError,
				Payload:    getError(strconv.Itoa(http.StatusInternalServerError), "Failed to delete CloudWatch audit trail"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusInternalServerError,
				Payload:    getError(strconv.Itoa(http.StatusInternalServerError), "Failed to delete Splunk audit trail"),
			},
		},
		{
			// Ensure Read is reachable for import negative-path tests
			// that need a 404 → nil result rather than a panic.
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/assets/dataSinks",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Not found"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/datasinks/splunk/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Not found"),
			},
		},
	}
}

// Silence unused-import complaint if strfmt is only used via other route
// files — this reference makes the import concrete regardless.
var _ = strfmt.DateTime{}
