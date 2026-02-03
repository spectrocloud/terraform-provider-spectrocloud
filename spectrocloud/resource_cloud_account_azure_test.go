package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// Test for toAzureAccount (table-driven)
func TestToAzureAccount_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		rdSet  map[string]interface{}
		verify func(t *testing.T, rd *schema.ResourceData, acc *models.V1AzureAccount)
	}{
		{
			name: "base account",
			rdSet: map[string]interface{}{
				"name": "azure_unit_test_acc", "context": "tenant",
				"azure_client_id": "test_client_id", "azure_client_secret": "test_client_secret",
				"azure_tenant_id": "test_tenant_id", "tenant_name": "test_tenant_name",
				"disable_properties_request": true, "private_cloud_gateway_id": "12345",
				"cloud": "AzureUSGovernmentCloud",
			},
			verify: func(t *testing.T, rd *schema.ResourceData, acc *models.V1AzureAccount) {
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
				assert.Nil(t, acc.Spec.TLS)
			},
		},
		{
			name: "with TLS cert",
			rdSet: map[string]interface{}{
				"name": "azure_unit_test_acc", "context": "tenant",
				"azure_client_id": "test_client_id", "azure_client_secret": "test_client_secret",
				"azure_tenant_id": "test_tenant_id", "tenant_name": "test_tenant_name",
				"disable_properties_request": true, "private_cloud_gateway_id": "12345",
				"cloud": "AzureUSSecretCloud", "tls_cert": "test-certificate-data",
			},
			verify: func(t *testing.T, rd *schema.ResourceData, acc *models.V1AzureAccount) {
				assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
				assert.Equal(t, "tenant", acc.Metadata.Annotations["scope"])
				assert.Equal(t, rd.Get("cloud"), *acc.Spec.AzureEnvironment)
				assert.NotNil(t, acc.Spec.TLS)
				assert.Equal(t, "test-certificate-data", acc.Spec.TLS.Cert)
			},
		},
		{
			name: "without TLS cert",
			rdSet: map[string]interface{}{
				"name": "azure_unit_test_acc", "context": "tenant",
				"azure_client_id": "test_client_id", "azure_client_secret": "test_client_secret",
				"azure_tenant_id": "test_tenant_id", "tenant_name": "test_tenant_name",
				"disable_properties_request": true, "private_cloud_gateway_id": "12345",
				"cloud": "AzurePublicCloud", "tls_cert": "",
			},
			verify: func(t *testing.T, rd *schema.ResourceData, acc *models.V1AzureAccount) {
				assert.Equal(t, rd.Get("name"), acc.Metadata.Name)
				assert.Equal(t, rd.Get("cloud"), *acc.Spec.AzureEnvironment)
				assert.Nil(t, acc.Spec.TLS)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountAzure().TestResourceData()
			for k, v := range tt.rdSet {
				rd.Set(k, v)
			}
			acc := toAzureAccount(rd)
			tt.verify(t, rd, acc)
		})
	}
}

// Test for flattenCloudAccountAzure (table-driven)
func TestFlattenCloudAccountAzure_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		account *models.V1AzureAccount
		expect  map[string]interface{}
	}{
		{
			name: "base account",
			account: &models.V1AzureAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "test_account",
					Annotations: map[string]string{OverlordUID: "12345"},
					UID:         "abcdef",
				},
				Spec: &models.V1AzureCloudAccount{
					ClientID:         types.Ptr("test_client_id"),
					ClientSecret:     types.Ptr("test_client_secret"),
					TenantID:         types.Ptr("test_tenant_id"),
					TenantName:       "test_tenant_name",
					Settings:         &models.V1CloudAccountSettings{DisablePropertiesRequest: true},
					AzureEnvironment: types.Ptr("AzureUSGovernmentCloud"),
				},
			},
			expect: map[string]interface{}{
				"name": "test_account", "private_cloud_gateway_id": "12345",
				"azure_client_id": "test_client_id", "azure_tenant_id": "test_tenant_id",
				"tenant_name": "test_tenant_name", "disable_properties_request": true,
				"cloud": "AzureUSGovernmentCloud",
			},
		},
		{
			name: "with TLS cert",
			account: &models.V1AzureAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "test_account",
					Annotations: map[string]string{OverlordUID: "12345"},
					UID:         "abcdef",
				},
				Spec: &models.V1AzureCloudAccount{
					ClientID:         types.Ptr("test_client_id"),
					ClientSecret:     types.Ptr("test_client_secret"),
					TenantID:         types.Ptr("test_tenant_id"),
					TenantName:       "test_tenant_name",
					Settings:         &models.V1CloudAccountSettings{DisablePropertiesRequest: true},
					AzureEnvironment: types.Ptr("AzureUSSecretCloud"),
					TLS:              &models.V1AzureSecretTLSConfig{Cert: "test-certificate-data"},
				},
			},
			expect: map[string]interface{}{
				"name": "test_account", "private_cloud_gateway_id": "12345",
				"azure_client_id": "test_client_id", "azure_tenant_id": "test_tenant_id",
				"tenant_name": "test_tenant_name", "disable_properties_request": true,
				"cloud": "AzureUSSecretCloud", "tls_cert": "test-certificate-data",
			},
		},
		{
			name: "without TLS cert",
			account: &models.V1AzureAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "test_account",
					Annotations: map[string]string{OverlordUID: "12345"},
					UID:         "abcdef",
				},
				Spec: &models.V1AzureCloudAccount{
					ClientID:         types.Ptr("test_client_id"),
					ClientSecret:     types.Ptr("test_client_secret"),
					TenantID:         types.Ptr("test_tenant_id"),
					TenantName:       "test_tenant_name",
					Settings:         &models.V1CloudAccountSettings{DisablePropertiesRequest: true},
					AzureEnvironment: types.Ptr("AzurePublicCloud"),
					TLS:              nil,
				},
			},
			expect: map[string]interface{}{
				"name": "test_account", "cloud": "AzurePublicCloud", "tls_cert": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountAzure().TestResourceData()
			diags, hasError := flattenCloudAccountAzure(rd, tt.account)
			assert.Nil(t, diags)
			assert.False(t, hasError)
			for k, want := range tt.expect {
				assert.Equal(t, want, rd.Get(k), "field %s", k)
			}
		})
	}
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

// Test Create/Update with invalid TLS cert configuration (table-driven)
func TestResourceCloudAccountAzureInvalidTlsCert_TableDriven(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name  string
		op    func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
		rdSet map[string]interface{}
		setID string
	}{
		{
			name: "Create with invalid TLS cert",
			op:   resourceCloudAccountAzureCreate,
			rdSet: map[string]interface{}{
				"name": "test-azure-account", "context": "project",
				"azure_tenant_id": "tenant-azure-id", "azure_client_id": "azure-client-id",
				"azure_client_secret": "test-client-secret", "cloud": "AzurePublicCloud",
				"tls_cert": "invalid-cert-for-public-cloud",
			},
		},
		{
			name: "Update with invalid TLS cert",
			op:   resourceCloudAccountAzureUpdate,
			rdSet: map[string]interface{}{
				"name": "test-azure-account", "context": "project",
				"azure_tenant_id": "tenant-azure-id", "azure_client_id": "azure-client-id",
				"azure_client_secret": "test-client-secret", "cloud": "AzureUSGovernmentCloud",
				"tls_cert": "invalid-cert-for-gov-cloud",
			},
			setID: "test-azure-account-id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := resourceCloudAccountAzure().TestResourceData()
			for k, v := range tt.rdSet {
				rd.Set(k, v)
			}
			if tt.setID != "" {
				rd.SetId(tt.setID)
			}
			diags := tt.op(ctx, rd, unitTestMockAPIClient)
			assert.Len(t, diags, 1)
			assert.True(t, diags.HasError())
			assert.Contains(t, diags[0].Summary, "tls_cert can only be set when cloud is 'AzureUSSecretCloud'")
		})
	}
}
