package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// Batch 5 adds Minio/GCP/Azure BSL endpoints so
// {Minio,Gcp,Azure}BackupStorageLocation{Create,Read,Update} in
// common_backup_storage_location.go can be exercised. The list endpoint
// (/v1/users/assets/locations) is extended to include one BSL of each
// storage type so GetBackupStorageLocation resolves the right UID→type
// mapping during Read.
func BackupRoutes() []Route {
	// UID convention used by the batch-5 tests. Kept as string
	// constants so both this mock and the tests reference identical
	// values.
	const (
		s3BSLUID    = "test-bsl-location-id"
		minioBSLUID = "test-minio-bsl-id"
		gcpBSLUID   = "test-gcp-bsl-id"
		azureBSLUID = "test-azure-bsl-id"
	)

	return []Route{
		// --- Validate endpoints ---
		{
			Method: "POST",
			Path:   "/v1/clouds/aws/s3/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/gcp/bucketname/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/azure/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},

		// --- S3 ---
		{
			Method: "POST",
			Path:   "/v1/users/assets/locations/s3",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-backup-location-id"},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/users/assets/locations/s3/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/locations/s3/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/locations/s3/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1UserAssetsLocationS3{
					Metadata: &models.V1ObjectMetaInputEntity{
						Annotations: nil,
						Labels:      nil,
						Name:        "test-backup-location",
					},
					Spec: &models.V1UserAssetsLocationS3Spec{
						Config: &models.V1S3StorageConfig{
							BucketName: strPtr("test-bucket"),
							CaCert:     "test-cert",
							Credentials: &models.V1AwsCloudAccount{
								AccessKey:      "test-access-key",
								CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
								Partition:      nil,
								PolicyARNs:     []string{"test-arn"},
								SecretKey:      "test-secret-key",
								Sts:            nil,
							},
							Region:           strPtr("test-east"),
							S3ForcePathStyle: boolPtr(false),
							S3URL:            "s3://test/test",
							UseRestic:        nil,
						},
						IsDefault: false,
						Type:      "",
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/locations",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1UserAssetsLocations{
					Items: []*models.V1UserAssetsLocation{
						{
							Metadata: &models.V1ObjectMeta{Name: "test-bsl-location", UID: s3BSLUID},
							Spec: &models.V1UserAssetsLocationSpec{
								Storage: models.V1LocationTypeS3.Pointer(),
							},
						},
						{
							Metadata: &models.V1ObjectMeta{Name: "test-minio-bsl", UID: minioBSLUID},
							Spec: &models.V1UserAssetsLocationSpec{
								Storage: models.V1LocationTypeMinio.Pointer(),
							},
						},
						{
							Metadata: &models.V1ObjectMeta{Name: "test-gcp-bsl", UID: gcpBSLUID},
							Spec: &models.V1UserAssetsLocationSpec{
								Storage: models.V1LocationTypeGcp.Pointer(),
							},
						},
						{
							Metadata: &models.V1ObjectMeta{Name: "test-azure-bsl", UID: azureBSLUID},
							Spec: &models.V1UserAssetsLocationSpec{
								// V1LocationType has no Azure constant in the current SDK;
								// AzureBackupStorageLocationRead doesn't switch on the
								// Storage type anyway — it always calls
								// GetAzureBackupStorageLocation. Leave a plain S3 pointer
								// so the outer GetBackupStorageLocation returns this row.
								Storage: models.V1LocationTypeS3.Pointer(),
							},
						},
					},
				},
			},
		},

		// --- Minio ---
		{
			Method: "POST",
			Path:   "/v1/users/assets/locations/minio",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": minioBSLUID},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/locations/minio/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/locations/minio/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1UserAssetsLocationS3{
					Metadata: &models.V1ObjectMetaInputEntity{Name: "test-minio-bsl"},
					Spec: &models.V1UserAssetsLocationS3Spec{
						Config: &models.V1S3StorageConfig{
							BucketName: strPtr("test-minio-bucket"),
							Region:     strPtr("us-east-1"),
							S3URL:      "https://minio.example.com",
							Credentials: &models.V1AwsCloudAccount{
								CredentialType: models.V1AwsCloudAccountCredentialTypeSecret.Pointer(),
								AccessKey:      "minio-access-key",
								SecretKey:      "minio-secret-key",
							},
							S3ForcePathStyle: boolPtr(true),
						},
					},
				},
			},
		},

		// --- GCP ---
		{
			Method: "POST",
			Path:   "/v1/users/assets/locations/gcp",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": gcpBSLUID},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/locations/gcp/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/locations/gcp/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1UserAssetsLocationGcp{
					Metadata: &models.V1ObjectMetaInputEntity{Name: "test-gcp-bsl"},
					Spec: &models.V1UserAssetsLocationGcpSpec{
						Config: &models.V1GcpStorageConfig{
							BucketName: strPtr("test-gcp-bucket"),
							ProjectID:  "test-project",
							Credentials: &models.V1GcpAccountEntitySpec{
								JSONCredentials: `{"type":"service_account"}`,
							},
						},
					},
				},
			},
		},

		// --- Azure ---
		{
			Method: "POST",
			Path:   "/v1/users/assets/locations/azure",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": azureBSLUID},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/users/assets/locations/azure/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/users/assets/locations/azure/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1UserAssetsLocationAzure{
					Metadata: &models.V1ObjectMetaInputEntity{Name: "test-azure-bsl"},
					Spec: &models.V1UserAssetsLocationAzureSpec{
						Config: &models.V1AzureStorageConfig{
							ContainerName: strPtr("test-container"),
							StorageName:   strPtr("teststorage"),
							Sku:           "Standard_LRS",
							ResourceGroup: strPtr("test-rg"),
							Credentials: &models.V1AzureAccountEntitySpec{
								ClientCloud:    strPtr("public"),
								ClientID:       "test-client-id",
								ClientSecret:   "test-client-secret",
								SubscriptionID: "test-sub-id",
								TenantID:       "test-tenant-id",
							},
						},
					},
				},
			},
		},
	}
}
