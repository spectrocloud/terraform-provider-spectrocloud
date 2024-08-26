package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test for the `toGcpAccount` function
func TestToGcpAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1GcpAccountEntity
	}{
		{
			name: "All Fields Present",
			input: map[string]interface{}{
				"name":                 "gcp-account",
				"gcp_json_credentials": "credentials-json",
			},
			expected: &models.V1GcpAccountEntity{
				Metadata: &models.V1ObjectMeta{
					Name: "gcp-account",
					UID:  "",
				},
				Spec: &models.V1GcpAccountEntitySpec{
					JSONCredentials: "credentials-json",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceCloudAccountGcp().Schema, tt.input)

			// Call the function under test
			result := toGcpAccount(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.UID, result.Metadata.UID)
			assert.Equal(t, tt.expected.Spec.JSONCredentials, result.Spec.JSONCredentials)
		})
	}
}
