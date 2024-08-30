package routes

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func getAccountResponse(cloud string) interface{} {
	switch cloud {
	case "aws":
		return &models.V1AwsAccounts{
			Items: []*models.V1AwsAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-aws-account-1",
						UID:  "test-aws-account-id-1",
					},
					Spec: &models.V1AwsCloudAccount{
						AccessKey:      "test-access-key",
						CredentialType: "secret",
						Partition:      nil,
						PolicyARNs:     nil,
						SecretKey:      "test-crt",
						Sts: &models.V1AwsStsCredentials{
							Arn:        "test-arn",
							ExternalID: "test-ex-id",
						},
					},
					Status: &models.V1CloudAccountStatus{State: "active"},
				},
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-aws-account-2",
						UID:  generateRandomStringUID(),
					},
					Spec: &models.V1AwsCloudAccount{
						AccessKey:      "test-access-key",
						CredentialType: "secret",
						Partition:      nil,
						PolicyARNs:     nil,
						SecretKey:      "test-crt",
						Sts: &models.V1AwsStsCredentials{
							Arn:        "test-arn",
							ExternalID: "test-ex-id",
						},
					},
					Status: &models.V1CloudAccountStatus{State: "active"},
				},
			},
			Listmeta: &models.V1ListMetaData{
				Continue: "",
				Count:    2,
				Limit:    10,
				Offset:   0,
			},
		}
	case "azure":
		return &models.V1AzureAccounts{
			Items: []*models.V1AzureAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-azure-account-1",
						UID:  "test-azure-account-id-1",
					},
					Spec: &models.V1AzureCloudAccount{
						AzureEnvironment: ptr.StringPtr("test-env"),
						ClientID:         ptr.StringPtr("test-client-id"),
						ClientSecret:     ptr.StringPtr("test-secret"),
						Settings:         nil,
						TenantID:         ptr.StringPtr("tenant-id"),
						TenantName:       "test",
					},
					Status: nil,
				},
			},
			Listmeta: nil,
		}
	case "tke":
		return &models.V1TencentAccounts{
			Items: []*models.V1TencentAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Annotations:           nil,
						CreationTimestamp:     models.V1Time{},
						DeletionTimestamp:     models.V1Time{},
						Labels:                nil,
						LastModifiedTimestamp: models.V1Time{},
						Name:                  "test-tke-account-1",
						UID:                   "test-tke-account-id-1",
					},
					Spec: &models.V1TencentCloudAccount{
						SecretID:  ptr.StringPtr("test-secretID"),
						SecretKey: ptr.StringPtr("test-secretKey"),
					},
					Status: &models.V1CloudAccountStatus{
						State: "active",
					},
				},
			},
			Listmeta: &models.V1ListMetaData{
				Continue: "",
				Count:    2,
				Limit:    10,
				Offset:   0,
			},
		}
	case "gcp":
		return &models.V1GcpAccounts{
			Items: []*models.V1GcpAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Annotations:           nil,
						CreationTimestamp:     models.V1Time{},
						DeletionTimestamp:     models.V1Time{},
						Labels:                nil,
						LastModifiedTimestamp: models.V1Time{},
						Name:                  "test-gcp-account-1",
						UID:                   "test-gcp-account-id-1",
					},
					Spec:   nil,
					Status: nil,
				},
			},
			Listmeta: nil,
		}
	case "vsphere":
		return &models.V1VsphereAccounts{
			Items: []*models.V1VsphereAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-vsphere-account-1",
						UID:  "test-vsphere-account-id-1",
					},
				},
			},
			Listmeta: nil,
		}
	case "openstack":
		return &models.V1OpenStackAccounts{
			Items: []*models.V1OpenStackAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-openstack-account-1",
						UID:  "test-openstack-account-id-1",
					},
				},
			},
			Listmeta: nil,
		}
	case "maas":
		return &models.V1MaasAccounts{
			Items: []*models.V1MaasAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-maas-account-1",
						UID:  "test-maas-account-id-1",
					},
					Spec: &models.V1MaasCloudAccount{
						APIEndpoint:      ptr.StringPtr("test.end.com"),
						APIKey:           ptr.StringPtr("testApiKey"),
						PreferredSubnets: []string{"subnet1"},
					},
				},
			},
			Listmeta: nil,
		}
	case "custom":
		return &models.V1CustomAccounts{
			Items: []*models.V1CustomAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-custom-account-1",
						UID:  "test-custom-account-id-1",
					},
				},
			},
			Listmeta: nil,
		}
	}
	return nil
}

func getAccountNegativeResponse(cloud string) interface{} {
	switch cloud {
	case "aws":
		return &models.V1AwsAccounts{
			Items: []*models.V1AwsAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-aws-account-2",
						UID:  generateRandomStringUID(),
					},
					Spec: &models.V1AwsCloudAccount{
						AccessKey:      "test-access-key",
						CredentialType: "secret",
						Partition:      nil,
						PolicyARNs:     nil,
						SecretKey:      "test-crt",
						Sts: &models.V1AwsStsCredentials{
							Arn:        "test-arn",
							ExternalID: "test-ex-id",
						},
					},
					Status: &models.V1CloudAccountStatus{State: "active"},
				},
			},
			Listmeta: &models.V1ListMetaData{
				Continue: "",
				Count:    2,
				Limit:    10,
				Offset:   0,
			},
		}
	case "azure":
		return &models.V1AzureAccounts{
			Items: []*models.V1AzureAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-azure-account-1-neg",
						UID:  "test-azure-account-id-1-neg",
					},
					Spec: &models.V1AzureCloudAccount{
						AzureEnvironment: ptr.StringPtr("test-env"),
						ClientID:         ptr.StringPtr("test-client-id"),
						ClientSecret:     ptr.StringPtr("test-secret"),
						Settings:         nil,
						TenantID:         ptr.StringPtr("tenant-id"),
						TenantName:       "test",
					},
					Status: nil,
				},
			},
			Listmeta: nil,
		}
	case "tke":
		return &models.V1TencentAccounts{
			Items: []*models.V1TencentAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Annotations:           nil,
						CreationTimestamp:     models.V1Time{},
						DeletionTimestamp:     models.V1Time{},
						Labels:                nil,
						LastModifiedTimestamp: models.V1Time{},
						Name:                  "test--1",
						UID:                   "test-id-1",
					},
					Spec: &models.V1TencentCloudAccount{
						SecretID:  ptr.StringPtr("test-secretID"),
						SecretKey: ptr.StringPtr("test-secretKey"),
					},
					Status: &models.V1CloudAccountStatus{
						State: "notActive",
					},
				},
			},
			Listmeta: &models.V1ListMetaData{
				Continue: "",
				Count:    2,
				Limit:    10,
				Offset:   0,
			},
		}
	case "gcp":
		return &models.V1GcpAccounts{
			Items: []*models.V1GcpAccount{
				{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Annotations:           nil,
						CreationTimestamp:     models.V1Time{},
						DeletionTimestamp:     models.V1Time{},
						Labels:                nil,
						LastModifiedTimestamp: models.V1Time{},
						Name:                  "test-gcp-1-neg",
						UID:                   "test-account-gcp-id-1-neg",
					},
					Spec:   nil,
					Status: nil,
				},
			},
			Listmeta: nil,
		}
	case "vsphere":
		return &models.V1VsphereAccounts{
			Items: []*models.V1VsphereAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-vsphere-account-1-neg",
						UID:  "test-vsphere-account-id-1-neg",
					},
				},
			},
			Listmeta: nil,
		}
	case "openstack":
		return &models.V1OpenStackAccounts{
			Items: []*models.V1OpenStackAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-openstack-account-1-neg",
						UID:  "test-openstack-account-uid-1-neg",
					},
				},
			},
			Listmeta: nil,
		}
	case "maas":
		return &models.V1MaasAccounts{
			Items: []*models.V1MaasAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-maas-account-1-neg",
						UID:  "test-maas-account-id-1-neg",
					},
					Spec: &models.V1MaasCloudAccount{
						APIEndpoint:      ptr.StringPtr("test.end.com"),
						APIKey:           ptr.StringPtr("testApiKey"),
						PreferredSubnets: []string{"subnet1"},
					},
				},
			},
			Listmeta: nil,
		}
	case "custom":
		return &models.V1CustomAccounts{
			Items: []*models.V1CustomAccount{
				{
					Metadata: &models.V1ObjectMeta{
						Name: "test-custom-account-1-neg",
						UID:  "test-custom-account-id-1-neg",
					},
				},
			},
			Listmeta: nil,
		}
	}
	return nil
}

func CloudAccountsRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/gcp",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("gcp"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/azure",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("azure"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/aws",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("aws"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/tencent",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("tke"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/vsphere",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("vsphere"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/openstack",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("openstack"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/maas",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("maas"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/cloudTypes/{cloudType}",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("custom"),
			},
		},
	}
}

func CloudAccountsNegativeRoutes() []Route {
	return []Route{
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/gcp",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("gcp"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/azure",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("azure"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/aws",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("aws"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/tencent",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("tke"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/vsphere",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("vsphere"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/openstack",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("openstack"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/maas",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("maas"),
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/cloudTypes/{cloudType}",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountNegativeResponse("custom"),
			},
		},
	}
}
