package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
	"strconv"
)

//func getMockMacroPayload() models.V1Macro {
//	return models.V1Macro{
//		Name:  "SampleMacro",
//		Value: "SampleValue",
//	}
//}

func getMockMacrosPayload() *models.V1Macros {
	return &models.V1Macros{
		Macros: []*models.V1Macro{
			{
				Name:  "macro1",
				Value: "value1",
			},
			{
				Name:  "macro2",
				Value: "value2",
			},
		},
	}
}

func MacrosRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},

		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockMacrosPayload(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockMacrosPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

func MacrosNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "Macro already exists"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "Macro not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusMethodNotAllowed,
				Payload:    getError(strconv.Itoa(http.StatusNoContent), "Operation not allowed"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/projects/{uid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "Macro not found"),
			},
		},
		// for tenant
		{
			Method: "POST",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "Macro already exists"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "Macro not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusMethodNotAllowed,
				Payload:    getError(strconv.Itoa(http.StatusNoContent), "Operation not allowed"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/tenants/{tenantUid}/macros",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "Macro not found"),
			},
		},
	}
}
