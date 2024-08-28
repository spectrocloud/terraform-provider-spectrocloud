package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
	"strconv"
)

func getFiltersResponse() models.V1FiltersSummary {
	return models.V1FiltersSummary{
		Items: []*models.V1FilterSummary{
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-filter-1",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1FilterSummarySpec{
					FilterType: "test",
				},
			},
			{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-filter-2",
					UID:                   generateRandomStringUID(),
				},
				Spec: &models.V1FilterSummarySpec{
					FilterType: "test",
				},
			},
		},
		Listmeta: nil,
	}
}

func FilterRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/filters",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getFiltersResponse(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/filters/tag/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getFiltersResponse(),
			},
		},
	}
}

func FilterNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/filters",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "filter not found"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/filters/tag/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "filter not found"),
			},
		},
	}
}
