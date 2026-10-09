package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func mockPublicCloudRateConfigPayload() *models.V1PublicCloudRateConfig {
	return &models.V1PublicCloudRateConfig{
		ComputeOptimized: &models.V1CloudInstanceRateConfig{
			ComputeRateProportion: 65,
			MemoryRateProportion:  35,
		},
		MemoryOptimized: &models.V1CloudInstanceRateConfig{
			ComputeRateProportion: 25,
			MemoryRateProportion:  75,
		},
	}
}

func mockPrivateCloudRateConfigPayload() *models.V1PrivateCloudRateConfig {
	return &models.V1PrivateCloudRateConfig{
		CPUUnitPricePerHour:        0.021811,
		GpuUnitPricePerHour:        2.933908,
		MemoryUnitPriceGiBPerHour:  0.002923,
		StorageUnitPriceGiBPerHour: 0.000028,
	}
}

// mockRateConfigPayload mirrors the real GET response, which always comes back
// with every cloud populated plus one entry per registered custom cloud.
func mockRateConfigPayload() *models.V1RateConfig {
	return &models.V1RateConfig{
		Aws:              mockPublicCloudRateConfigPayload(),
		Azure:            mockPublicCloudRateConfigPayload(),
		Gcp:              mockPublicCloudRateConfigPayload(),
		Vsphere:          mockPrivateCloudRateConfigPayload(),
		Maas:             mockPrivateCloudRateConfigPayload(),
		Generic:          mockPrivateCloudRateConfigPayload(),
		ApacheCloudstack: mockPrivateCloudRateConfigPayload(),
		Custom: []*models.V1CustomCloudRateConfig{
			{
				CloudType:  "test-custom-cloud",
				RateConfig: mockPrivateCloudRateConfigPayload(),
			},
		},
	}
}

// RateConfigRoutes wires the two SDK endpoints that resource_rate_config.go and
// data_source_rate_config.go touch:
//   - /v1/tenants/{tenantUid}/rateConfig (GET/PUT)
func RateConfigRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/rateConfig",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockRateConfigPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/rateConfig",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

// RateConfigNegativeRoutes returns error responses on both paths. Unlike the
// developer setting and password policy wrappers, GetRateConfig checks err
// before dereferencing the payload, so a failing GET is safe to exercise.
func RateConfigNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/tenants/{tenantUid}/rateConfig",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid rate config"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/tenants/{tenantUid}/rateConfig",
			Response: ResponseData{
				StatusCode: http.StatusBadRequest,
				Payload:    getError(strconv.Itoa(http.StatusBadRequest), "Invalid rate config"),
			},
		},
	}
}
