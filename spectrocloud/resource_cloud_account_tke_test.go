package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
					SecretID:  types.Ptr("test-secret-id"),
					SecretKey: types.Ptr("test-secret-key"),
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
					SecretID:  types.Ptr(""),
					SecretKey: types.Ptr(""),
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
