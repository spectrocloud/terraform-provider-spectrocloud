package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func WorkSpaceRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/workspaces",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-ws-1"},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/workspaces/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/workspaces/{uid}",
			Response: ResponseData{
				StatusCode: 0,
				Payload: &models.V1Workspace{
					Metadata: &models.V1ObjectMeta{
						Annotations: nil,
						Labels:      nil,
						Name:        "test-ws-1",
						UID:         "test-ws-1-id",
					},
					Spec: &models.V1WorkspaceSpec{
						ClusterNamespaces: []*models.V1WorkspaceClusterNamespace{
							{
								Image: &models.V1WorkspaceNamespaceImage{
									BlackListedImages: []string{"image1"},
								},
								IsRegex: false,
								Name:    "test-ws-ns",
								NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
									ClusterResourceAllocations: []*models.V1ClusterResourceAllocation{
										{
											ClusterUID: "test-cluster-uid",
											ResourceAllocation: &models.V1WorkspaceResourceAllocation{
												CPUCores:  2,
												MemoryMiB: 100,
											},
										},
									},
									DefaultResourceAllocation: &models.V1WorkspaceResourceAllocation{
										CPUCores:  2,
										MemoryMiB: 100,
									},
								},
							},
						},
						ClusterRbacs: []*models.V1ClusterRbac{
							{
								Metadata: &models.V1ObjectMeta{
									Name: "test-rbac-name",
									UID:  "test-rbac-id",
								},
								Spec: &models.V1ClusterRbacSpec{
									Bindings: []*models.V1ClusterRbacBinding{
										{
											Namespace: "test-ns",
											Role:      nil,
											Subjects:  nil,
											Type:      "ns",
										},
									},
									RelatedObject: &models.V1RelatedObject{
										Kind: "test",
										Name: "test-ro",
										UID:  "test-ro-id",
									},
								},
								Status: &models.V1ClusterRbacStatus{
									Errors: nil,
								},
							},
						},
						ClusterRefs: []*models.V1WorkspaceClusterRef{
							{
								ClusterName: "test-cluster-name",
								ClusterUID:  "test-cluster-id",
							},
						},
						Policies: &models.V1WorkspacePolicies{
							BackupPolicy: &models.V1WorkspaceBackupConfigEntity{
								BackupConfig: &models.V1ClusterBackupConfig{
									BackupLocationName:      "test-bl",
									BackupLocationUID:       "uid",
									BackupName:              "test-back-name",
									BackupPrefix:            "prefix",
									DurationInHours:         0,
									IncludeAllDisks:         false,
									IncludeClusterResources: false,
									LocationType:            "test-location",
									Namespaces:              nil,
									Schedule:                nil,
								},
								ClusterUids:        []string{"c-uid"},
								IncludeAllClusters: false,
							},
						},
						Quota: &models.V1WorkspaceQuota{
							ResourceAllocation: &models.V1WorkspaceResourceAllocation{
								CPUCores:  2,
								MemoryMiB: 100,
							},
						},
					},
					Status: &models.V1WorkspaceStatus{
						Errors: nil,
					},
				},
			},
		},
	}
}
