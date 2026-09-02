package routes

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// ResponseData defines the structure of mock responses
type ResponseData struct {
	StatusCode int
	Payload    interface{}
}

// Route defines a mock route with method, path, and response.
//
// A route may specify either a static Response, or a dynamic Handler
// (http.HandlerFunc) that lets a single path emit different payloads across
// calls — useful for expressing Create → Read after-create → Update → Read
// after-update state transitions that a fixed payload can't capture.
//
// When Handler is set it takes precedence over Response; otherwise the
// server writes Response.StatusCode + JSON-encoded Response.Payload.
type Route struct {
	Method   string
	Path     string
	Response ResponseData
	Handler  http.HandlerFunc
}

func generateRandomStringUID() string {
	bytes := make([]byte, 24/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "test"
	}
	return hex.EncodeToString(bytes)
}

func CommonProjectRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/health",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: map[string]interface{}{
					"healthy": true,
				},
			},
		},
	}
}

// strPtr and boolPtr return pointers to the given value. They exist so mock
// route files don't need to import the spectrocloud package (which would
// create an import cycle now that mockserver is imported by the spectrocloud
// package's own tests). They intentionally mirror spectrocloud.StringPtr /
// spectrocloud.BoolPtr signatures — call-sites are drop-in replaceable.
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func getError(code string, msg string) models.V1Error {
	return models.V1Error{
		Code:    code,
		Details: nil,
		Message: msg,
		Ref:     "ref-" + generateRandomStringUID(),
	}
}
