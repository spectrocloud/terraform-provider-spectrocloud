package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
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

func getFilterSummary() *models.V1TagFilterSummary {
	return &models.V1TagFilterSummary{
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "test-filter-2",
			UID:                   generateRandomStringUID(),
		},
		Spec: &models.V1TagFilterSpec{
			FilterGroup: &models.V1TagFilterGroup{
				Conjunction: (*models.V1SearchFilterConjunctionOperator)(ptr.To("and")),
				Filters: []*models.V1TagFilterItem{
					{
						Key:      "name",
						Negation: false,
						Operator: "",
						Values:   []string{"test"},
					},
				},
			},
		},
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
			Method: "POST",
			Path:   "/v1/filters/tag",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-filter-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/filters/tag/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/filters/tag/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/filters/tag/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getFilterSummary(),
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
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "filter not found"),
			},
		},
	}
}
