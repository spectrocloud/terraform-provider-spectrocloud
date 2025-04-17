package routes

import "github.com/spectrocloud/palette-sdk-go/api/models"

func ClusterGroupRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/clustergroups",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-cg-1"},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/clustergroups/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/clustergroups/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clustergroups/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1ClusterGroup{
					Metadata: &models.V1ObjectMeta{
						Annotations: nil,
						Labels: map[string]string{
							"test": "dev",
						},
						Name: "test-cg",
						UID:  "test-cg-1",
					},
					Spec: &models.V1ClusterGroupSpec{
						ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
							{
								CloudType:        "aws",
								Name:             "temp1",
								PackServerRefs:   nil,
								PackServerSecret: "test-secret",
								Packs:            nil,
								ProfileVersion:   "1.0.0",
								RelatedObject:    nil,
								Type:             "cluster",
								UID:              "test-uid",
								Version:          0,
							},
						},
						ClusterRefs: []*models.V1ClusterGroupClusterRef{
							{
								ClusterName: "test-cluster",
								ClusterUID:  "test-cluster-id",
							},
						},
						ClustersConfig: &models.V1ClusterGroupClustersConfig{
							EndpointType: "test-end",
							HostClustersConfig: []*models.V1ClusterGroupHostClusterConfig{
								{
									ClusterUID: "test-cluster-id",
									EndpointConfig: &models.V1HostClusterEndpointConfig{
										IngressConfig: &models.V1IngressConfig{
											Host: "121.0.0.1",
											Port: 1001,
										},
										LoadBalancerConfig: &models.V1LoadBalancerConfig{
											ExternalIPs:              []string{"0.0.0.0"},
											ExternalTrafficPolicy:    "policy",
											LoadBalancerSourceRanges: []string{"0.0.0.1"},
										},
									},
								},
							},
							KubernetesDistroType: models.V1ClusterKubernetesDistroTypeCncfK8s.Pointer(),
							LimitConfig:          nil,
							Values:               "test-values",
						},
						Type: "",
					},
					Status: &models.V1ClusterGroupStatus{
						IsActive: true,
					},
				},
			},
		},
	}
}
