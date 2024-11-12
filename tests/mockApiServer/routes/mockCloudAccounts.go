package routes

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
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
						AzureEnvironment: ptr.To("test-env"),
						ClientID:         ptr.To("test-client-id"),
						ClientSecret:     ptr.To("test-secret"),
						Settings:         nil,
						TenantID:         ptr.To("tenant-id"),
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
						SecretID:  ptr.To("test-secretID"),
						SecretKey: ptr.To("test-secretKey"),
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
						APIEndpoint:      ptr.To("test.end.com"),
						APIKey:           ptr.To("testApiKey"),
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
						AzureEnvironment: ptr.To("test-env"),
						ClientID:         ptr.To("test-client-id"),
						ClientSecret:     ptr.To("test-secret"),
						Settings:         nil,
						TenantID:         ptr.To("tenant-id"),
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
						SecretID:  ptr.To("test-secretID"),
						SecretKey: ptr.To("test-secretKey"),
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
						APIEndpoint:      ptr.To("test.end.com"),
						APIKey:           ptr.To("testApiKey"),
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
			Path:   "/v1/cloudaccounts/summary",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1CloudAccountsSummary{
					Items: []*models.V1CloudAccountSummary{
						{
							Kind: "",
							Metadata: &models.V1ObjectMeta{
								Annotations: map[string]string{"scope": "project"},
								Name:        "test-import-account",
								UID:         "test-import-acc-id",
							},
							SpecSummary: &models.V1CloudAccountSummarySpecSummary{
								AccountID: "test-import-acc-id",
							},
							Status: &models.V1CloudAccountStatus{
								State: "Active",
							},
						},
					},
					Listmeta: nil,
				},
			},
		},

		// gcp
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/gcp",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-gcp-account-id-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/gcp/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/gcp/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/gcp/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/gcp/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1GcpAccount{
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
					Spec: &models.V1GcpAccountSpec{
						JSONCredentials:         "test-json-cred",
						JSONCredentialsFileName: "test-json",
					},
					Status: &models.V1CloudAccountStatus{
						State: "Running",
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/overlords/gcp/{uid}/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/gcp",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("gcp"),
			},
		},

		// Maas
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/maas",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-maas-account-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/maas/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/maas/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/maas/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
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
			Path:   "/v1/cloudaccounts/maas/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1MaasAccount{
					Metadata: &models.V1ObjectMeta{
						Name:        "test-maas-account-1",
						UID:         "test-maas-account-id-1",
						Annotations: map[string]string{"overlordUid": "test-pcg-id"},
					},
					Spec: &models.V1MaasCloudAccount{
						APIEndpoint:      ptr.To("test.end.com"),
						APIKey:           ptr.To("testApiKey"),
						PreferredSubnets: []string{"subnet1"},
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/overlords/maas/{uid}/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},

		// azure
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/azure",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-aws-account-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/azure/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/azure/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/azure/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/azure/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1AzureAccount{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Annotations: map[string]string{"scope": "project", "overlordUid": ""},
						Labels:      nil,
						Name:        "test-azure-account-1",
						UID:         "test-azure-account-id-1",
					},
					Spec: &models.V1AzureCloudAccount{
						AzureEnvironment: ptr.To("test-env"),
						ClientID:         ptr.To("test-client-id"),
						ClientSecret:     ptr.To("test-secret"),
						Settings: &models.V1CloudAccountSettings{
							DisablePropertiesRequest: false,
						},
						TenantID:   ptr.To("tenant-id"),
						TenantName: "test",
					},
					Status: nil,
				},
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

		// aws
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/aws",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("aws"),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/aws",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-aws-account-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/aws/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/aws/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/aws/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/aws/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1AwsAccount{
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
			},
		},

		// tke
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/tencent",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("tke"),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/tencent",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-tke-account-id-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/tencent/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/tencent/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/tencent/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/tencent/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1TencentAccount{
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
						SecretID:  ptr.To("test-secretID"),
						SecretKey: ptr.To("test-secretKey"),
					},
					Status: &models.V1CloudAccountStatus{
						State: "active",
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/overlords/tencent/{uid}/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},

		// vsphere
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/vsphere",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("vsphere"),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/vsphere",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-vsphere-account-id-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/vsphere/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/vsphere/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/vsphere/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/vsphere/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1VsphereAccount{

					Metadata: &models.V1ObjectMeta{
						Name: "test-vsphere-account-1",
						UID:  "test-vsphere-account-id-1",
					},
					Spec: &models.V1VsphereCloudAccount{
						Insecure:      false,
						Password:      ptr.To("test-pwd"),
						Username:      ptr.To("test-uname"),
						VcenterServer: ptr.To("test-uname.com"),
					},
					Status: &models.V1CloudAccountStatus{
						State: "Running",
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/overlords/vsphere/{uid}/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},

		// openstack
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/openstack",
			Response: ResponseData{
				StatusCode: 200,
				Payload:    getAccountResponse("openstack"),
			},
		},
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/openstack",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "test-openstack-account-id-1"},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/clouds/openstack/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/openstack/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/openstack/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/openstack/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1OpenStackAccount{
					Metadata: &models.V1ObjectMeta{
						Name: "test-openstack-account-1",
						UID:  "test-openstack-account-id-1",
					},
					Spec: &models.V1OpenStackCloudAccount{
						CaCert:           "testcert",
						DefaultDomain:    "test.com",
						DefaultProject:   "Default",
						IdentityEndpoint: ptr.To("testtest"),
						Insecure:         false,
						ParentRegion:     "test-region",
						Password:         ptr.To("test-pwd"),
						Username:         ptr.To("test-uname"),
					},
					Status: &models.V1CloudAccountStatus{
						State: "Running",
					},
				},
			},
		},
		{
			Method: "POST",
			Path:   "/v1/overlords/openstack/{uid}/account/validate",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    map[string]string{"AuditUID": generateRandomStringUID()},
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
		{
			Method: "POST",
			Path:   "/v1/cloudaccounts/cloudTypes/{cloudType}",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "mock-uid"},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/clouds/cloudTypes",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1CustomCloudTypes{
					CloudTypes: []*models.V1CustomCloudType{
						{
							CloudCategory: "test",
							CloudFamily:   "",
							DisplayName:   "test-cloud",
							IsCustom:      true,
							IsManaged:     false,
							IsVertex:      false,
							Logo:          "",
							Name:          "test-cloud",
						},
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/overlords/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1Overlord{
					Kind: "",
					Metadata: &models.V1ObjectMeta{
						Name: "pcg-1",
						UID:  "pcg-1-id",
					},
					Spec: &models.V1OverloadSpec{
						CloudAccountUID:   "test-acc-id",
						IPAddress:         "121.0.0.1",
						IPPools:           nil,
						IsSelfHosted:      false,
						IsSystem:          false,
						SpectroClusterUID: "test-spectro-id",
						TenantUID:         "test-tenant-id",
					},
					Status: &models.V1OverloadStatus{
						Health:          nil,
						IsActive:        false,
						IsReady:         false,
						KubectlCommands: nil,
						Notifications:   nil,
						State:           "Running",
					},
				},
			},
		},
		{
			Method: "GET",
			Path:   "/v1/cloudaccounts/cloudTypes/{cloudType}/{uid}",
			Response: ResponseData{
				StatusCode: 200,
				Payload: &models.V1CustomAccount{
					APIVersion: "",
					Kind:       "",
					Metadata: &models.V1ObjectMeta{
						Name: "test-name",
						UID:  "test-uid",
					},
					Spec: &models.V1CustomCloudAccount{
						Credentials: map[string]string{
							"username": "test",
							"password": "test",
						},
					},
					Status: &models.V1CloudAccountStatus{
						State: "Active",
					},
				},
			},
		},
		{
			Method: "PUT",
			Path:   "/v1/cloudaccounts/cloudTypes/{cloudType}/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
			},
		},
		{
			Method: "DELETE",
			Path:   "/v1/cloudaccounts/cloudTypes/{cloudType}/{uid}",
			Response: ResponseData{
				StatusCode: 204,
				Payload:    nil,
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
