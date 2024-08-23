package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
					APIEndpoint: types.Ptr("http://api.endpoint"),
					APIKey:      types.Ptr("api-key"),
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