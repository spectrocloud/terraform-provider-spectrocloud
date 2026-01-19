package routes

import (
	"net/http"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud"
)

func getHelmRegistryPayload() *models.V1HelmRegistry {
	return &models.V1HelmRegistry{
		APIVersion: "",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Annotations:           nil,
			CreationTimestamp:     models.V1Time{},
			DeletionTimestamp:     models.V1Time{},
			Labels:                nil,
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "Public",
			UID:                   "test-registry-uid",
		},
		Spec: &models.V1HelmRegistrySpec{
			Auth: &models.V1RegistryAuth{
				Password: "test=pwd",
				TLS:      nil,
				Token:    "as",
				Type:     "token",
				Username: "sf",
			},
			Endpoint:    spectrocloud.StringPtr("test.com"),
			IsPrivate:   false,
			Name:        "Public",
			RegistryUID: generateRandomStringUID(),
			Scope:       "project",
		},
		Status: &models.V1HelmRegistryStatus{
			HelmSyncStatus: &models.V1RegistrySyncStatus{
				LastRunTime:    models.V1Time{},
				LastSyncedTime: models.V1Time{},
				Message:        "",
				Status:         "Active",
			},
		},
	}
}

func RegistriesRoutes() []Route {
	return []Route{
		{
			Method: "PUT",
			Path:   "/v1/registries/oci/{uid}/ecr",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/registries/oci/{uid}/ecr",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/oci/{uid}/ecr",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1EcrRegistry{
					Kind: "",
					Metadata: &models.V1ObjectMeta{
						Annotations:           nil,
						CreationTimestamp:     models.V1Time{},
						DeletionTimestamp:     models.V1Time{},
						Labels:                nil,
						LastModifiedTimestamp: models.V1Time{},
						Name:                  "testSecretRegistry",
						UID:                   "testSecretRegistry-id",
					},
					Spec: &models.V1EcrRegistrySpec{
						BaseContentPath: "test-path",
						Credentials: &models.V1AwsCloudAccount{
							AccessKey:      "test-key",
							CredentialType: models.V1AwsCloudAccountCredentialTypeSts.Pointer(),
							Partition:      spectrocloud.StringPtr("test-part"),
							PolicyARNs:     []string{"test-arns"},
							SecretKey:      "test-secret-key",
							Sts: &models.V1AwsStsCredentials{
								Arn:        "test-arn",
								ExternalID: "test-external-id",
							},
						},
						DefaultRegion: "test-region",
						Endpoint:      spectrocloud.StringPtr("test.point"),
						IsPrivate:     spectrocloud.BoolPtr(false),
						ProviderType:  spectrocloud.StringPtr("test-type"),
						RegistryUID:   "test-reg-uid",
						Scope:         "project",
						TLS: &models.V1TLSConfiguration{
							Ca:                 "test-ca",
							Certificate:        "test-cert",
							Enabled:            false,
							InsecureSkipVerify: false,
							Key:                "test-key",
						},
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/ecr",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-sts-oci-reg-ecr-uid"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/ecr/validate",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/basic/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/oci/basic",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-zarf-oci-reg-basic-uid"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/oci/{uid}/basic",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1BasicOciRegistry{
					Kind: "",
					Metadata: &models.V1ObjectMeta{
						Annotations:           nil,
						CreationTimestamp:     models.V1Time{},
						DeletionTimestamp:     models.V1Time{},
						Labels:                nil,
						LastModifiedTimestamp: models.V1Time{},
						Name:                  "test-zarf-registry",
						UID:                   "test-zarf-oci-reg-basic-uid",
					},
					Spec: &models.V1BasicOciRegistrySpec{
						Endpoint:        spectrocloud.StringPtr("https://registry.example.com"),
						BasePath:        "",
						BaseContentPath: "/",
						ProviderType:    spectrocloud.StringPtr("zarf"),
						IsSyncSupported: true,
						Auth: &models.V1RegistryAuth{
							Username: "test-username",
							Password: "test-password",
							Type:     "basic",
							TLS: &models.V1TLSConfiguration{
								Certificate:        "",
								Enabled:            false,
								InsecureSkipVerify: false,
							},
						},
						Scope: "tenant",
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/registries/oci/{uid}/basic",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/registries/oci/{uid}/basic",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/oci/summary",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1OciRegistries{
					Items: []*models.V1OciRegistry{
						{
							Metadata: &models.V1ObjectMeta{
								Name: "test-registry-oci",
								UID:  "test-registry-uid",
							},
							Spec:   nil,
							Status: nil,
						},
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/registries/helm",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": generateRandomStringUID()},
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1HelmRegistries{
					Items: []*models.V1HelmRegistry{getHelmRegistryPayload()},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getHelmRegistryPayload(),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}/sync/status",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1RegistrySyncStatus{
					IsSyncSupported: true,
					Status:          "Success",
					Message:         "Registry synchronized successfully",
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/metadata",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1RegistriesMetadata{
					Items: []*models.V1RegistryMetadata{
						{
							IsDefault: false,
							IsPrivate: false,
							Kind:      "",
							Name:      "test-registry-name",
							Scope:     "project",
							UID:       "test-registry-uid",
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/registries/pack",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload: &models.V1PackRegistries{
					Items: []*models.V1PackRegistry{
						{
							APIVersion: "",
							Kind:       "",
							Metadata: &models.V1ObjectMeta{
								Annotations:           nil,
								CreationTimestamp:     models.V1Time{},
								DeletionTimestamp:     models.V1Time{},
								Labels:                nil,
								LastModifiedTimestamp: models.V1Time{},
								Name:                  "test-registry-name",
								UID:                   "test-registry-uid",
							},
							Spec:   nil,
							Status: nil,
						},
					},
					Listmeta: nil,
				},
			},
		},
	}
}

func RegistriesNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/registries/helm/{uid}",
			Response: ResponseData{
				StatusCode: http.StatusOK,
				Payload:    getHelmRegistryPayload(),
			},
		},
	}
}
