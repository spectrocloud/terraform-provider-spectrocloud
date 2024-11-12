package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

// Test for the `toMaasAccount` function
func TestToMaasAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1MaasAccount
	}{
		{
			name: "All Fields Present",
			input: map[string]interface{}{
				"name":                     "maas-account",
				"private_cloud_gateway_id": "private-cloud-gateway-id",
				"maas_api_endpoint":        "http://api.endpoint",
				"maas_api_key":             "api-key",
			},
			expected: &models.V1MaasAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "maas-account",
					Annotations: map[string]string{OverlordUID: "private-cloud-gateway-id"},
					UID:         "",
				},
				Spec: &models.V1MaasCloudAccount{
					APIEndpoint: ptr.To("http://api.endpoint"),
					APIKey:      ptr.To("api-key"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceCloudAccountMaas().Schema, tt.input)

			// Call the function under test
			result := toMaasAccount(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.Annotations[OverlordUID], result.Metadata.Annotations[OverlordUID])
			if tt.expected.Spec.APIEndpoint == nil {
				assert.Nil(t, result.Spec.APIEndpoint)
			} else {
				assert.Equal(t, tt.expected.Spec.APIEndpoint, result.Spec.APIEndpoint)
			}
			if tt.expected.Spec.APIKey == nil {
				assert.Nil(t, result.Spec.APIKey)
			} else {
				assert.Equal(t, tt.expected.Spec.APIKey, result.Spec.APIKey)
			}
		})
	}
}

func prepareResourceCloudAccountMaas() *schema.ResourceData {
	d := resourceCloudAccountMaas().TestResourceData()
	d.SetId("test-maas-account-id-1")
	_ = d.Set("name", "test-maas-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "test-pcg-id")
	_ = d.Set("maas_api_endpoint", "test-maas-api-endpoint")
	_ = d.Set("maas_api_key", "test-maas-api-key")
	return d
}
func TestResourceCloudAccountMaasCreate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountMaas()
	ctx := context.Background()
	diags := resourceCloudAccountMaasCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-maas-account-1", d.Id())
}

func TestResourceCloudAccountMaasRead(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountMaas()
	ctx := context.Background()
	diags := resourceCloudAccountMaasRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-maas-account-id-1", d.Id())
}
func TestResourceCloudAccountMaasUpdate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountMaas()
	ctx := context.Background()
	diags := resourceCloudAccountMaasUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-maas-account-id-1", d.Id())
}
func TestResourceCloudAccountMaasDelete(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountMaas()
	ctx := context.Background()
	diags := resourceCloudAccountMaasDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)

}
