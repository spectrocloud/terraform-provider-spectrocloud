package routes

import (
	"bytes"
	v1 "github.com/spectrocloud/palette-sdk-go/api/client/version1"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func ClusterRoutes() []Route {
	var buffer bytes.Buffer
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/dashboard/spectroclusters/search",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1SpectroClustersSummary{
					Items: []*models.V1SpectroClusterSummary{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-cluster",
								UID:  "test-cluster-id",
							},
							SpecSummary: nil,
							Status:      nil,
						},
					},
					Listmeta: &models.V1ListMetaData{
						Continue: "",
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1SpectroCluster{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-cluster",
						UID:  "test-cluster-id",
					},
					Spec: nil,
					Status: &models.V1SpectroClusterStatus{

						State: "Running",
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/assets/kubeconfig",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &v1.V1SpectroClustersUIDKubeConfigOK{
					ContentDisposition: "test-content",
					Payload:            &buffer,
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/assets/adminKubeconfig",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &v1.V1SpectroClustersUIDKubeConfigOK{
					ContentDisposition: "test-content",
					Payload:            &buffer,
				},
			},
		},
	}
}
