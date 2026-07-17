package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// mockPCGDNSMapPayload is the fixture returned by the positive GET route.
func mockPCGDNSMapPayload() *models.V1VsphereDNSMapping {
	dc := "dc1"
	name := "example.com"
	network := "vm-network"
	pcg := "test-pcg-uid"
	return &models.V1VsphereDNSMapping{
		Metadata: &models.V1ObjectMeta{
			Name: name,
			UID:  "test-dnsmap-uid",
		},
		Spec: &models.V1VsphereDNSMappingSpec{
			Datacenter:        &dc,
			DNSName:           &name,
			Network:           &network,
			PrivateGatewayUID: &pcg,
		},
	}
}

// PCGDNSMapRoutes returns the happy-path route set for the PCG DNS map
// resource so resource_pcg_dns_map CRUD tests can exercise Create/Read/
// Update/Delete against the in-process mock server.
func PCGDNSMapRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/users/assets/vsphere/dnsMappings",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    &models.V1UID{UID: strPtr("test-dnsmap-uid")},
			},
		},
		{
			// List endpoint for data_source_pcg_dns_map (Batch 17). The
			// SDK issues GET /vsphere/dnsMappings?filter=... and expects
			// a paginated list envelope; return one matching item.
			Method: "GET",
			Path:   "/v1/users/assets/vsphere/dnsMappings",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1VsphereDNSMappings{
					Items: []*models.V1VsphereDNSMapping{mockPCGDNSMapPayload()},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/vsphere/dnsMappings/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    mockPCGDNSMapPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/vsphere/dnsMappings/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/users/assets/vsphere/dnsMappings/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}
