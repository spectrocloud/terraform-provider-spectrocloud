package routes

import (
	"bytes"
	"net/http"

	v1 "github.com/spectrocloud/palette-sdk-go/api/client/version1"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud"
)

func getMockSpectroCluster() *models.V1SpectroCluster {
	return &models.V1SpectroCluster{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Name: "test-cluster",
			UID:  "test-cluster-id",
			Labels: map[string]string{
				"env": "test",
			},
		},
		Spec: &models.V1SpectroClusterSpec{
			ClusterType: "full",
			ClusterConfig: &models.V1ClusterConfig{
				ClusterMetaAttribute:        "test-cluster-meta-attributes",
				UpdateWorkerPoolsInParallel: true,
				Timezone:                    "UTC",
			},
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  clusterProfileUID1,
					Name: "test-cluster-profile-1",
					Type: "cluster",
				},
				{
					UID:  clusterProfileUID2,
					Name: "test-cluster-profile-2",
					Type: "addon",
				},
			},
		},
		Status: &models.V1SpectroClusterStatus{
			State: "Running",
			SpcApply: &models.V1SpcApply{
				CanBeApplied: true,
			},
		},
	}
}

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
				Payload:    getMockSpectroCluster(),
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
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/assets/kubeconfigclient",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &v1.V1SpectroClustersUIDKubeConfigClientGetOK{
					ContentDisposition: "test-content",
					Payload:            &buffer,
				},
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/spectroclusters/{uid}/profiles",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/spectroclusters/{uid}/profiles",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/variables",
			Response: ResponseData{
				StatusCode: 200,
				Payload: []*models.V1SpectroClusterVariables{
					{
						ProfileUID: spectrocloud.StringPtr(clusterProfileUID1),
						Variables: []*models.V1SpectroClusterVariableResponse{
							{
								Name:  spectrocloud.StringPtr("region"),
								Value: "us-east-1",
							},
						},
					},
				},
			},
		},
		{
			Method: "PATCH",
			Path:   "/v1/spectroclusters/{uid}/variables",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/features/backup",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterBackup{
					Spec: &models.V1ClusterBackupSpec{},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/features/complianceScan",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterComplianceScan{
					Spec: &models.V1ClusterComplianceScanSpec{},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/config/rbacs",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterRbacs{
					Items: []*models.V1ClusterRbac{},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/spectroclusters/{uid}/config/namespaces",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterNamespaceResources{
					Items: []*models.V1ClusterNamespaceResource{},
				},
			},
		},
	}
}

func ClusterNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/dashboard/spectroclusters/search",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1SpectroClustersSummary{
					Items:    []*models.V1SpectroClusterSummary{},
					Listmeta: &models.V1ListMetaData{},
				},
			},
		},
	}
}
