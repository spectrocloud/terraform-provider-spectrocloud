package routes

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
)

// ResponseData defines the structure of mock responses
type ResponseData struct {
	StatusCode int
	Payload    interface{}
}

// Route defines a mock route with method, path, and response
type Route struct {
	Method   string
	Path     string
	Response ResponseData
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

func getError(code string, msg string) models.V1Error {
	return models.V1Error{
		Code:    code,
		Details: nil,
		Message: msg,
		Ref:     "ref-" + generateRandomStringUID(),
	}
}
