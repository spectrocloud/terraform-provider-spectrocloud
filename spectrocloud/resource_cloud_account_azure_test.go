package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
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
			ClientID:     ptr.To("test_client_id"),
			ClientSecret: ptr.To("test_client_secret"),
			TenantID:     ptr.To("test_tenant_id"),
			TenantName:   "test_tenant_name",
			Settings: &models.V1CloudAccountSettings{
				DisablePropertiesRequest: true,
			},
			AzureEnvironment: ptr.To("AzureUSGovernmentCloud"),
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
	return d
}

func TestResourceCloudAccountAzureCreate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountAzureTestData()
	ctx := context.Background()
	diags := resourceCloudAccountAzureCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-aws-account-1", d.Id())
}

func TestResourceCloudAccountAzureRead(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountAzureTestData()
	ctx := context.Background()
	diags := resourceCloudAccountAzureRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-azure-account-id-1", d.Id())
}

func TestResourceCloudAccountAzureUpdate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountAzureTestData()
	ctx := context.Background()
	diags := resourceCloudAccountAzureUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-azure-account-id-1", d.Id())
}

func TestResourceCloudAccountAzureDelete(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountAzureTestData()
	ctx := context.Background()
	diags := resourceCloudAccountAzureDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
