package routes

import (
	"net/http"
	"strconv"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func getEdgeHostSearchSummary() models.V1EdgeHostsSearchSummary {
	var items []*models.V1EdgeHostsMetadata
	var profileSummary []*models.V1ProfileTemplateSummary
	profileSummary = append(profileSummary, &models.V1ProfileTemplateSummary{
		CloudType: "aws",
		Name:      "test-profile-1",
		Packs: []*models.V1PackRefSummary{{
			AddonType:   "",
			Annotations: nil,
			DisplayName: "k8",
			Layer:       "infra",
			LogoURL:     "",
			Name:        "kubernetes_pack",
			PackUID:     generateRandomStringUID(),
			Tag:         "",
			Type:        "",
			Version:     "1.28.0",
		}},
		Type:    "cluster",
		UID:     generateRandomStringUID(),
		Version: "1.0",
	})
	items = append(items, &models.V1EdgeHostsMetadata{
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "test-edge-01",
			UID:                   generateRandomStringUID(),
		},
		Spec: &models.V1EdgeHostsMetadataSpec{
			ClusterProfileTemplates: profileSummary,
			Device: &models.V1DeviceSpec{
				ArchType: ptr.To("AMD"),
				CPU: &models.V1CPU{
					Cores: 2,
				},
				Disks: []*models.V1Disk{{
					Controller: "",
					Partitions: nil,
					Size:       50,
					Vendor:     "",
				}},
				Gpus: []*models.V1GPUDeviceSpec{
					{
						Addresses: map[string]string{
							"test": "121.0.0.1",
						},
						Model:  "xyz",
						Vendor: "abc",
					},
				},
				Memory: nil,
				Nics:   nil,
				Os:     nil,
			},
			Host: &models.V1EdgeHostSpecHost{
				HostAddress: "192.168.1.100",
				MacAddress:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			ProjectMeta: nil,
			Type:        "",
		},
		Status: &models.V1EdgeHostsMetadataStatus{
			Health: &models.V1EdgeHostHealth{
				AgentVersion: "",
				Message:      "",
				State:        "healthy",
			},
			InUseClusters: nil,
			State:         "",
		},
	})
	return models.V1EdgeHostsSearchSummary{
		Items: items,
		Listmeta: &models.V1ListMetaData{
			Continue: "",
			Count:    1,
			Limit:    50,
			Offset:   0,
		},
	}
}

func getEdgeHostPayload() models.V1EdgeHostDevice {
	return models.V1EdgeHostDevice{
		Aclmeta: &models.V1ACLMeta{
			OwnerUID:   generateRandomStringUID(),
			ProjectUID: generateRandomStringUID(),
			TenantUID:  generateRandomStringUID(),
		},
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                map[string]string{"type": "test"},
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "test-edge-01",
			UID:                   generateRandomStringUID(),
		},
		Spec: &models.V1EdgeHostDeviceSpec{
			CloudProperties:         nil,
			ClusterProfileTemplates: nil,
			Device: &models.V1DeviceSpec{
				ArchType: ptr.To("amd64"),
				CPU:      nil,
				Disks:    nil,
				Gpus:     nil,
				Memory:   nil,
				Nics:     nil,
				Os:       nil,
			},
			Host:       nil,
			Properties: nil,
			Service:    nil,
			Type:       "",
			Version:    "1.0",
		},
		Status: &models.V1EdgeHostDeviceStatus{
			Health: &models.V1EdgeHostHealth{
				AgentVersion: "",
				Message:      "",
				State:        "healthy",
			},
			InUseClusters:    nil,
			Packs:            nil,
			ProfileStatus:    nil,
			ServiceAuthToken: "",
			State:            "ready",
		},
	}
}

//func creatEdgeHostErrorResponse() interface{} {
//	var payload interface{}
//	payload = map[string]interface{}{
//		"UID": ptr.To("test-edge-host-id"),
//	}
//	return map[string]interface{}{
//		"AuditUID": generateRandomStringUID(),
//		"Payload":  payload,
//	}
//}

func AppliancesRoutes() []Route {
	return []Route{
		{
			Method: "DELETE",
			Path:   "/v1/edgehosts/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload: map[string]string{
					"err": "test_error",
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/dashboard/edgehosts/search",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getEdgeHostSearchSummary(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/edgehosts/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getEdgeHostPayload(),
			},
		},
	}
}

func AppliancesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/edgehosts",
			Response: ResponseData{
				StatusCode: http.StatusLocked,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "Operation not allowed"),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/dashboard/edgehosts/search",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "No edge host found"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/edgehosts/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "No edge host found"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/edgehosts/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusNotFound), "No edge host found"),
			},
		},
	}
}
