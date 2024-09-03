package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"net/http"
	"strconv"
)

func getMockWorkspacePayload() *models.V1Workspace {
	return &models.V1Workspace{
		Metadata: &models.V1ObjectMeta{
			Name: "Default",
			UID:  generateRandomStringUID(),
			Annotations: map[string]string{
				"description": "An example workspace for testing",
			},
			Labels: map[string]string{
				"env": "test",
			},
		},
		Spec: &models.V1WorkspaceSpec{
			ClusterNamespaces: []*models.V1WorkspaceClusterNamespace{
				{
					Image:   nil,
					IsRegex: false,
					Name:    "Default",
					NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
						ClusterResourceAllocations: []*models.V1ClusterResourceAllocation{
							{
								ClusterUID:         generateRandomStringUID(),
								ResourceAllocation: nil,
							},
						},
						DefaultResourceAllocation: &models.V1WorkspaceResourceAllocation{
							CPUCores:  1000,
							MemoryMiB: 4,
						},
					},
				},
			},
			ClusterRbacs: []*models.V1ClusterRbac{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "Default",
						UID:  generateRandomStringUID(),
					},
					Spec: &models.V1ClusterRbacSpec{
						Bindings: []*models.V1ClusterRbacBinding{
							{
								Namespace: "Default",
								Type:      "DefaultType",
								Role: &models.V1ClusterRoleRef{
									Name: "Default",
									Kind: "DefaultKind",
								},
							},
						},
						RelatedObject: &models.V1RelatedObject{
							Kind: "DefaultKind",
							Name: "Default",
							UID:  generateRandomStringUID(),
						},
					},
					Status: &models.V1ClusterRbacStatus{
						Errors: []*models.V1ClusterResourceError{},
					},
				},
			},
			ClusterRefs: []*models.V1WorkspaceClusterRef{
				{
					ClusterUID: generateRandomStringUID(),
				},
			},
			Policies: &models.V1WorkspacePolicies{
				BackupPolicy: &models.V1WorkspaceBackupConfigEntity{
					BackupConfig: &models.V1ClusterBackupConfig{
						// Add relevant fields with dummy data for BackupConfig here
						Schedule: &models.V1ClusterFeatureSchedule{
							ScheduledRunTime: "daily",
						},
						BackupLocationName: "Default", // Keep backups for 7 days
					},
					ClusterUids:        []string{generateRandomStringUID()}, // Dummy cluster UIDs
					IncludeAllClusters: true,
				},
			},
			Quota: &models.V1WorkspaceQuota{
				ResourceAllocation: &models.V1WorkspaceResourceAllocation{
					// Add relevant fields with dummy data here
					CPUCores:  1000,
					MemoryMiB: 4,
				},
			},
		},
		Status: &models.V1WorkspaceStatus{
			Errors: []*models.V1WorkspaceError{},
		},
	}
}

func getMockWorkspaceBackUpPayload() *models.V1WorkspaceBackup {
	return &models.V1WorkspaceBackup{
		Metadata: &models.V1ObjectMeta{
			Name:        "Deafult-backup",
			UID:         generateRandomStringUID(),
			Labels:      map[string]string{"environment": "dev"},
			Annotations: map[string]string{"createdBy": "testUser"},
		},
		Spec: &models.V1WorkspaceBackupSpec{
			// Populate with dummy data for the Spec
			Config: &models.V1WorkspaceBackupConfig{
				BackupConfig: &models.V1ClusterBackupConfig{
					// Add relevant fields with dummy data for BackupConfig here
					Schedule: &models.V1ClusterFeatureSchedule{
						ScheduledRunTime: "daily",
					},
					BackupLocationName: "Default", // Keep backups for 7 days
				},
				ClusterUids: []string{generateRandomStringUID()}, // Dummy cluster UIDs
			},
			WorkspaceUID: generateRandomStringUID(), // Keep backups for 7 days
		},
		Status: &models.V1WorkspaceBackupStatus{},
	}
}

func WorkspaceRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/workspaces",
			Response: ResponseData{
				StatusCode: http.StatusCreated,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/workspaces/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockWorkspacePayload(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/workspaces/{uid}/backup",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getMockWorkspaceBackUpPayload(),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/workspaces/{uid}/clusterNamespaces",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    map[string]interface{}{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/workspaces/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNoContent,
				Payload:    nil,
			},
		},
	}
}

func WorkspaceNegativeRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/workspaces",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "workspaces already exist"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/workspaces/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusForbidden,
				Payload:    getError(strconv.Itoa(http.StatusOK), "workspaces not found"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/workspaces/{uid}/backup",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "backup not found"),
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/workspaces/{uid}/clusterNamespaces",
			Response: ResponseData{
				StatusCode: http.StatusMethodNotAllowed,
				Payload:    getError(strconv.Itoa(http.StatusNoContent), "Operation not allowed"),
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/workspaces/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusNotFound,
				Payload:    getError(strconv.Itoa(http.StatusOK), "workspaces not found"),
			},
		},
	}
}

//func WorkSpaceRoutes() []Route {
//	return []Route{
//		{
//			Method: "POST",
//			Path:   "/v1/workspaces",
//			Response: ResponseData{
//				StatusCode: 201,
//				Payload:    map[string]string{"UID": "test-ws-1"},
//			},
//		},
//		{
//			Method: "DELETE",
//			Path:   "/v1/workspaces/{uid}",
//			Response: ResponseData{
//				StatusCode: 204,
//				Payload:    nil,
//			},
//		},
//		{
//			Method: "GET",
//			Path:   "/v1/workspaces/{uid}",
//			Response: ResponseData{
//				StatusCode: 0,
//				Payload: &models.V1Workspace{
//					Metadata: &models.V1ObjectMeta{
//						Annotations: nil,
//						Labels:      nil,
//						Name:        "test-ws-1",
//						UID:         "test-ws-1-id",
//					},
//					Spec: &models.V1WorkspaceSpec{
//						ClusterNamespaces: []*models.V1WorkspaceClusterNamespace{
//							{
//								Image: &models.V1WorkspaceNamespaceImage{
//									BlackListedImages: []string{"image1"},
//								},
//								IsRegex: false,
//								Name:    "test-ws-ns",
//								NamespaceResourceAllocation: &models.V1WorkspaceNamespaceResourceAllocation{
//									ClusterResourceAllocations: []*models.V1ClusterResourceAllocation{
//										{
//											ClusterUID: "test-cluster-uid",
//											ResourceAllocation: &models.V1WorkspaceResourceAllocation{
//												CPUCores:  2,
//												MemoryMiB: 100,
//											},
//										},
//									},
//									DefaultResourceAllocation: &models.V1WorkspaceResourceAllocation{
//										CPUCores:  2,
//										MemoryMiB: 100,
//									},
//								},
//							},
//						},
//						ClusterRbacs: []*models.V1ClusterRbac{
//							{
//								Metadata: &models.V1ObjectMeta{
//									Name: "test-rbac-name",
//									UID:  "test-rbac-id",
//								},
//								Spec: &models.V1ClusterRbacSpec{
//									Bindings: []*models.V1ClusterRbacBinding{
//										{
//											Namespace: "test-ns",
//											Role:      nil,
//											Subjects:  nil,
//											Type:      "ns",
//										},
//									},
//									RelatedObject: &models.V1RelatedObject{
//										Kind: "test",
//										Name: "test-ro",
//										UID:  "test-ro-id",
//									},
//								},
//								Status: &models.V1ClusterRbacStatus{
//									Errors: nil,
//								},
//							},
//						},
//						ClusterRefs: []*models.V1WorkspaceClusterRef{
//							{
//								ClusterName: "test-cluster-name",
//								ClusterUID:  "test-cluster-id",
//							},
//						},
//						Policies: &models.V1WorkspacePolicies{
//							BackupPolicy: &models.V1WorkspaceBackupConfigEntity{
//								BackupConfig: &models.V1ClusterBackupConfig{
//									BackupLocationName:      "test-bl",
//									BackupLocationUID:       "uid",
//									BackupName:              "test-back-name",
//									BackupPrefix:            "prefix",
//									DurationInHours:         0,
//									IncludeAllDisks:         false,
//									IncludeClusterResources: false,
//									LocationType:            "test-location",
//									Namespaces:              nil,
//									Schedule:                nil,
//								},
//								ClusterUids:        []string{"c-uid"},
//								IncludeAllClusters: false,
//							},
//						},
//						Quota: &models.V1WorkspaceQuota{
//							ResourceAllocation: &models.V1WorkspaceResourceAllocation{
//								CPUCores:  2,
//								MemoryMiB: 100,
//							},
//						},
//					},
//					Status: &models.V1WorkspaceStatus{
//						Errors: nil,
//					},
//				},
//			},
//		},
//	}
//}
