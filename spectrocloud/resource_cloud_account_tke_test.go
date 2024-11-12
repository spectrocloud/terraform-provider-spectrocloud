package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

// Test for the `toTencentAccount` function
func TestToTencentAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1TencentAccount
	}{
		{
			name: "Valid Data",
			input: map[string]interface{}{
				"name":               "tencent-account",
				"tencent_secret_id":  "test-secret-id",
				"tencent_secret_key": "test-secret-key",
			},
			expected: &models.V1TencentAccount{
				Metadata: &models.V1ObjectMeta{
					Name: "tencent-account",
					UID:  "", // UID is set from d.Id(), which is usually populated during resource creation
				},
				Spec: &models.V1TencentCloudAccount{
					SecretID:  ptr.To("test-secret-id"),
					SecretKey: ptr.To("test-secret-key"),
				},
			},
		},
		{
			name: "Empty Secret ID and Key",
			input: map[string]interface{}{
				"name":               "tencent-account",
				"tencent_secret_id":  "",
				"tencent_secret_key": "",
			},
			expected: &models.V1TencentAccount{
				Metadata: &models.V1ObjectMeta{
					Name: "tencent-account",
					UID:  "",
				},
				Spec: &models.V1TencentCloudAccount{
					SecretID:  ptr.To(""),
					SecretKey: ptr.To(""),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceCloudAccountTencent().Schema, tt.input)

			// Call the function under test
			result := toTencentAccount(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.UID, result.Metadata.UID)
			assert.Equal(t, *tt.expected.Spec.SecretID, *result.Spec.SecretID)
			assert.Equal(t, *tt.expected.Spec.SecretKey, *result.Spec.SecretKey)
		})
	}
}

func prepareResourceCloudAccountTencent() *schema.ResourceData {
	d := resourceCloudAccountTencent().TestResourceData()
	d.SetId("test-tke-account-id-1")
	_ = d.Set("name", "test-tke-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("tencent_secret_id", "test-secret-id")
	_ = d.Set("tencent_secret_key", "test-secret-key")

	return d
}

func TestResourceCloudAccountTencentCreate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountTencent()
	ctx := context.Background()
	diags := resourceCloudAccountTencentCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-tke-account-id-1", d.Id())
}

func TestResourceCloudAccountTencentRead(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountTencent()
	ctx := context.Background()
	diags := resourceCloudAccountTencentRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-tke-account-id-1", d.Id())
}

func TestResourceCloudAccountTencentUpdate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountTencent()
	ctx := context.Background()
	diags := resourceCloudAccountTencentUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-tke-account-id-1", d.Id())
}

func TestResourceCloudAccountTencentDelete(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountTencent()
	ctx := context.Background()
	diags := resourceCloudAccountTencentDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
