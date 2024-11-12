package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func BackupRoutes() []Route {
	return []Route{
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
							BucketName: ptr.To("test-bucket"),
							CaCert:     "test-cert",
							Credentials: &models.V1AwsCloudAccount{
								AccessKey:      "test-access-key",
								CredentialType: "secret",
								Partition:      nil,
								PolicyARNs:     []string{"test-arn"},
								SecretKey:      "test-secret-key",
								Sts:            nil,
							},
							Region:           ptr.To("test-east"),
							S3ForcePathStyle: ptr.To(false),
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
							Metadata: &models.V1ObjectMeta{
								Annotations: nil,
								Labels:      nil,
								Name:        "test-bsl-location",
								UID:         "test-bsl-location-id",
							},
							Spec: &models.V1UserAssetsLocationSpec{
								IsDefault: false,
								Storage:   "s3",
								Type:      "",
							},
						},
					},
				},
			},
		},
		{
			Method:   "",
			Path:     "",
			Response: ResponseData{},
		},
	}
}
