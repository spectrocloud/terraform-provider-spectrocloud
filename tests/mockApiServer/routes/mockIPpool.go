package routes

import "github.com/spectrocloud/palette-sdk-go/api/models"

func IPPoolRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/overlords/vsphere/{uid}/pools",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-pcg-id"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/overlords/vsphere/{uid}/pools/{poolUid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/overlords/vsphere/{uid}/pools/{poolUid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/overlords/vsphere/{uid}/pools",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1IPPools{
					Items: []*models.V1IPPoolEntity{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-name",
								UID:  "test-pcg-id",
							},
							Spec: &models.V1IPPoolEntitySpec{
								Pool: &models.V1Pool{
									End:     "test-end",
									Gateway: "test-gateway",
									Nameserver: &models.V1Nameserver{
										Addresses: []string{"test-address"},
										Search:    []string{"test-search"},
									},
									Prefix: 0,
									Start:  "teat-start",
									Subnet: "test-subnet",
								},
								PriavetGatewayUID:       "test-pcg-id",
								RestrictToSingleCluster: false,
							},
							Status: &models.V1IPPoolStatus{
								AllottedIps:        nil,
								AssociatedClusters: nil,
								InUse:              false,
							},
						},
					},
				},
			},
		},
	}
}
