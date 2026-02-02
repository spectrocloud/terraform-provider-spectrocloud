package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// Test for toAzureAccount
func TestToAzureAccount(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData() // Assuming this method exists
	rd.Set("name", "azure_unit_test_acc")
	rd.Set("context", "tenant")
	rd.Set("azure_client_id", "test_client_id")
	rd.Set("azure_client_secret", "test_client_secret")
	rd.Set("azure_tenant_id", "test_tenant_id")
	rd.Set("tenant_name", "test_tenant_name")
	rd.Set("disable_properties_request", true)
	rd.Set("private_cloud_gateway_id", "12345")
	rd.Set("cloud", "AzureUSGovernmentCloud")
	acc := toAzureAccount(rd)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("azure_client_id"), *acc.Spec.ClientID)
	assert.Equal(t, rd.Get("azure_client_secret"), *acc.Spec.ClientSecret)
	assert.Equal(t, rd.Get("azure_tenant_id"), *acc.Spec.TenantID)
	assert.Equal(t, rd.Get("tenant_name"), acc.Spec.TenantName)
	assert.Equal(t, rd.Get("disable_properties_request"), acc.Spec.Settings.DisablePropertiesRequest)
	assert.Equal(t, rd.Get("private_cloud_gateway_id"), acc.Metadata.Annotations[OverlordUID])
	assert.Equal(t, rd.Get("cloud"), *acc.Spec.AzureEnvironment)
	assert.Equal(t, rd.Id(), acc.Metadata.UID)
}

// Test for flattenCloudAccountAzure
func TestFlattenCloudAccountAzure(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData() // Assuming this method exists
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{
			Name:        "test_account",
			Annotations: map[string]string{OverlordUID: "12345"},
			UID:         "abcdef",
		},
		Spec: &models.V1AzureCloudAccount{
			ClientID:     types.Ptr("test_client_id"),
			ClientSecret: types.Ptr("test_client_secret"),
			TenantID:     types.Ptr("test_tenant_id"),
			TenantName:   "test_tenant_name",
			Settings: &models.V1CloudAccountSettings{
				DisablePropertiesRequest: true,
			},
			AzureEnvironment: types.Ptr("AzureUSGovernmentCloud"),
		},
	}

	diags, hasError := flattenCloudAccountAzure(rd, account)

	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "test_account", rd.Get("name"))
	assert.Equal(t, "12345", rd.Get("private_cloud_gateway_id"))
	assert.Equal(t, "test_client_id", rd.Get("azure_client_id"))
	assert.Equal(t, "test_tenant_id", rd.Get("azure_tenant_id"))
	assert.Equal(t, "test_tenant_name", rd.Get("tenant_name"))
	assert.Equal(t, true, rd.Get("disable_properties_request"))
	assert.Equal(t, "AzureUSGovernmentCloud", rd.Get("cloud"))
}

func prepareResourceCloudAccountAzureTestData() *schema.ResourceData {
	d := resourceCloudAccountAzure().TestResourceData()
	d.SetId("test-azure-account-id-1")
	_ = d.Set("name", "test-azure-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("azure_tenant_id", "tenant-azure-id")
	_ = d.Set("azure_client_id", "azure-client-id")
	_ = d.Set("azure_client_secret", "test-client-secret")
	_ = d.Set("tenant_name", "azure-tenant")
	_ = d.Set("disable_properties_request", false)
	_ = d.Set("cloud", "AzurePublicCloud")
	_ = d.Set("private_cloud_gateway_id", "test-pcg-id")
	return d
}

func TestResourceCloudAccountAzureCRUD(t *testing.T) {
	testResourceCRUD(t, prepareResourceCloudAccountAzureTestData, unitTestMockAPIClient,
		resourceCloudAccountAzureCreate, resourceCloudAccountAzureRead, resourceCloudAccountAzureUpdate, resourceCloudAccountAzureDelete)
}

// Test for validateTlsCertConfiguration function
func TestValidateTlsCertConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		cloud       string
		tlsCert     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid: AzureUSSecretCloud with tls_cert",
			cloud:       "AzureUSSecretCloud",
			tlsCert:     "test-certificate-data",
			expectError: false,
		},
		{
			name:        "Valid: AzureUSSecretCloud without tls_cert",
			cloud:       "AzureUSSecretCloud",
			tlsCert:     "",
			expectError: false,
		},
		{
			name:        "Valid: AzurePublicCloud without tls_cert",
			cloud:       "AzurePublicCloud",
			tlsCert:     "",
			expectError: false,
		},
		{
			name:        "Invalid: AzurePublicCloud with tls_cert",
			cloud:       "AzurePublicCloud",
			tlsCert:     "test-certificate-data",
			expectError: true,
			errorMsg:    "tls_cert can only be set when cloud is 'AzureUSSecretCloud', but cloud is set to 'AzurePublicCloud'",
		},
		{
			name:        "Invalid: AzureUSGovernmentCloud with tls_cert",
			cloud:       "AzureUSGovernmentCloud",
			tlsCert:     "test-certificate-data",
			expectError: true,
			errorMsg:    "tls_cert can only be set when cloud is 'AzureUSSecretCloud', but cloud is set to 'AzureUSGovernmentCloud'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountAzure().TestResourceData()
			rd.Set("cloud", tt.cloud)
			rd.Set("tls_cert", tt.tlsCert)

			err := validateTlsCertConfiguration(rd)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test for toAzureAccount with TLS certificate
func TestToAzureAccountWithTlsCert(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	rd.Set("name", "azure_unit_test_acc")
	rd.Set("context", "tenant")
	rd.Set("azure_client_id", "test_client_id")
	rd.Set("azure_client_secret", "test_client_secret")
	rd.Set("azure_tenant_id", "test_tenant_id")
	rd.Set("tenant_name", "test_tenant_name")
	rd.Set("disable_properties_request", true)
	rd.Set("private_cloud_gateway_id", "12345")
	rd.Set("cloud", "AzureUSSecretCloud")
	rd.Set("tls_cert", "test-certificate-data")

	acc := toAzureAccount(rd)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
	assert.Equal(t, rd.Get("azure_client_id"), *acc.Spec.ClientID)
	assert.Equal(t, rd.Get("azure_client_secret"), *acc.Spec.ClientSecret)
	assert.Equal(t, rd.Get("azure_tenant_id"), *acc.Spec.TenantID)
	assert.Equal(t, rd.Get("tenant_name"), acc.Spec.TenantName)
	assert.Equal(t, rd.Get("disable_properties_request"), acc.Spec.Settings.DisablePropertiesRequest)
	assert.Equal(t, rd.Get("private_cloud_gateway_id"), acc.Metadata.Annotations[OverlordUID])
	assert.Equal(t, rd.Get("cloud"), *acc.Spec.AzureEnvironment)
	assert.Equal(t, rd.Id(), acc.Metadata.UID)
	// Test TLS configuration
	assert.NotNil(t, acc.Spec.TLS)
	assert.Equal(t, "test-certificate-data", acc.Spec.TLS.Cert)
}

// Test for toAzureAccount without TLS certificate
func TestToAzureAccountWithoutTlsCert(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	rd.Set("name", "azure_unit_test_acc")
	rd.Set("context", "tenant")
	rd.Set("azure_client_id", "test_client_id")
	rd.Set("azure_client_secret", "test_client_secret")
	rd.Set("azure_tenant_id", "test_tenant_id")
	rd.Set("tenant_name", "test_tenant_name")
	rd.Set("disable_properties_request", true)
	rd.Set("private_cloud_gateway_id", "12345")
	rd.Set("cloud", "AzurePublicCloud")
	rd.Set("tls_cert", "") // Empty TLS cert

	acc := toAzureAccount(rd)

	assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
	assert.Equal(t, rd.Get("cloud"), *acc.Spec.AzureEnvironment)
	// Test TLS configuration should be nil when tls_cert is empty
	assert.Nil(t, acc.Spec.TLS)
}

// Test for flattenCloudAccountAzure with TLS certificate
func TestFlattenCloudAccountAzureWithTlsCert(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{
			Name:        "test_account",
			Annotations: map[string]string{OverlordUID: "12345"},
			UID:         "abcdef",
		},
		Spec: &models.V1AzureCloudAccount{
			ClientID:     types.Ptr("test_client_id"),
			ClientSecret: types.Ptr("test_client_secret"),
			TenantID:     types.Ptr("test_tenant_id"),
			TenantName:   "test_tenant_name",
			Settings: &models.V1CloudAccountSettings{
				DisablePropertiesRequest: true,
			},
			AzureEnvironment: types.Ptr("AzureUSSecretCloud"),
			TLS: &models.V1AzureSecretTLSConfig{
				Cert: "test-certificate-data",
			},
		},
	}

	diags, hasError := flattenCloudAccountAzure(rd, account)

	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "test_account", rd.Get("name"))
	assert.Equal(t, "12345", rd.Get("private_cloud_gateway_id"))
	assert.Equal(t, "test_client_id", rd.Get("azure_client_id"))
	assert.Equal(t, "test_tenant_id", rd.Get("azure_tenant_id"))
	assert.Equal(t, "test_tenant_name", rd.Get("tenant_name"))
	assert.Equal(t, true, rd.Get("disable_properties_request"))
	assert.Equal(t, "AzureUSSecretCloud", rd.Get("cloud"))
	assert.Equal(t, "test-certificate-data", rd.Get("tls_cert"))
}

// Test for flattenCloudAccountAzure without TLS certificate
func TestFlattenCloudAccountAzureWithoutTlsCert(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	account := &models.V1AzureAccount{
		Metadata: &models.V1ObjectMeta{
			Name:        "test_account",
			Annotations: map[string]string{OverlordUID: "12345"},
			UID:         "abcdef",
		},
		Spec: &models.V1AzureCloudAccount{
			ClientID:     types.Ptr("test_client_id"),
			ClientSecret: types.Ptr("test_client_secret"),
			TenantID:     types.Ptr("test_tenant_id"),
			TenantName:   "test_tenant_name",
			Settings: &models.V1CloudAccountSettings{
				DisablePropertiesRequest: true,
			},
			AzureEnvironment: types.Ptr("AzurePublicCloud"),
			TLS:              nil, // No TLS config
		},
	}

	diags, hasError := flattenCloudAccountAzure(rd, account)

	assert.Nil(t, diags)
	assert.False(t, hasError)
	assert.Equal(t, "test_account", rd.Get("name"))
	assert.Equal(t, "AzurePublicCloud", rd.Get("cloud"))
	// tls_cert should not be set when TLS config is nil
	assert.Equal(t, "", rd.Get("tls_cert"))
}

// Test Create function with invalid TLS cert configuration
func TestResourceCloudAccountAzureCreateWithInvalidTlsCert(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	rd.Set("name", "test-azure-account")
	rd.Set("context", "project")
	rd.Set("azure_tenant_id", "tenant-azure-id")
	rd.Set("azure_client_id", "azure-client-id")
	rd.Set("azure_client_secret", "test-client-secret")
	rd.Set("cloud", "AzurePublicCloud")
	rd.Set("tls_cert", "invalid-cert-for-public-cloud") // This should fail validation

	ctx := context.Background()
	diags := resourceCloudAccountAzureCreate(ctx, rd, unitTestMockAPIClient)

	assert.Len(t, diags, 1)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "tls_cert can only be set when cloud is 'AzureUSSecretCloud'")
}

// Test Update function with invalid TLS cert configuration
func TestResourceCloudAccountAzureUpdateWithInvalidTlsCert(t *testing.T) {
	rd := resourceCloudAccountAzure().TestResourceData()
	rd.SetId("test-azure-account-id")
	rd.Set("name", "test-azure-account")
	rd.Set("context", "project")
	rd.Set("azure_tenant_id", "tenant-azure-id")
	rd.Set("azure_client_id", "azure-client-id")
	rd.Set("azure_client_secret", "test-client-secret")
	rd.Set("cloud", "AzureUSGovernmentCloud")
	rd.Set("tls_cert", "invalid-cert-for-gov-cloud") // This should fail validation

	ctx := context.Background()
	diags := resourceCloudAccountAzureUpdate(ctx, rd, unitTestMockAPIClient)

	assert.Len(t, diags, 1)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "tls_cert can only be set when cloud is 'AzureUSSecretCloud'")
}
