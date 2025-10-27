package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
)

// Test for the `toCloudStackAccount` function
func TestToCloudStackAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1CloudStackAccount
	}{
		{
			name: "All Fields Present",
			input: map[string]interface{}{
				"name":                     "cloudstack-account",
				"private_cloud_gateway_id": "private-cloud-gateway-id",
				"api_url":                  "https://cloudstack.example.com:8080/client/api",
				"api_key":                  "api-key",
				"secret_key":               "secret-key",
				"ca_certificate":           "ca-cert-content",
				"insecure":                 false,
			},
			expected: &models.V1CloudStackAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "cloudstack-account",
					Annotations: map[string]string{OverlordUID: "private-cloud-gateway-id"},
					UID:         "",
				},
				Spec: &models.V1CloudStackCloudAccount{
					APIURL:    types.Ptr("https://cloudstack.example.com:8080/client/api"),
					APIKey:    types.Ptr("api-key"),
					SecretKey: types.Ptr("secret-key"),
					CaCert:    "ca-cert-content",
					Insecure:  false,
				},
			},
		},
		{
			name: "Insecure Mode Enabled",
			input: map[string]interface{}{
				"name":                     "cloudstack-insecure",
				"private_cloud_gateway_id": "pcg-id",
				"api_url":                  "https://cloudstack.example.com:8080/client/api",
				"api_key":                  "test-key",
				"secret_key":               "test-secret",
				"insecure":                 true,
			},
			expected: &models.V1CloudStackAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "cloudstack-insecure",
					Annotations: map[string]string{OverlordUID: "pcg-id"},
					UID:         "",
				},
				Spec: &models.V1CloudStackCloudAccount{
					APIURL:    types.Ptr("https://cloudstack.example.com:8080/client/api"),
					APIKey:    types.Ptr("test-key"),
					SecretKey: types.Ptr("test-secret"),
					CaCert:    "",
					Insecure:  true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceCloudAccountCloudStack().Schema, tt.input)

			// Call the function under test
			result := toCloudStackAccount(d)

			// Perform assertions
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.Annotations[OverlordUID], result.Metadata.Annotations[OverlordUID])

			if tt.expected.Spec.APIURL == nil {
				assert.Nil(t, result.Spec.APIURL)
			} else {
				assert.Equal(t, *tt.expected.Spec.APIURL, *result.Spec.APIURL)
			}

			if tt.expected.Spec.APIKey == nil {
				assert.Nil(t, result.Spec.APIKey)
			} else {
				assert.Equal(t, *tt.expected.Spec.APIKey, *result.Spec.APIKey)
			}

			if tt.expected.Spec.SecretKey == nil {
				assert.Nil(t, result.Spec.SecretKey)
			} else {
				assert.Equal(t, *tt.expected.Spec.SecretKey, *result.Spec.SecretKey)
			}

			assert.Equal(t, tt.expected.Spec.CaCert, result.Spec.CaCert)
			assert.Equal(t, tt.expected.Spec.Insecure, result.Spec.Insecure)
		})
	}
}

func prepareResourceCloudAccountCloudStack() *schema.ResourceData {
	d := resourceCloudAccountCloudStack().TestResourceData()
	d.SetId("test-cloudstack-account-id-1")
	_ = d.Set("name", "test-cloudstack-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "test-pcg-id")
	_ = d.Set("api_url", "https://test.cloudstack.com:8080/client/api")
	_ = d.Set("api_key", "test-api-key")
	_ = d.Set("secret_key", "test-secret-key")
	_ = d.Set("insecure", false)
	return d
}

func TestResourceCloudAccountCloudStackCreate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountCloudStackCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cloudstack-account-1", d.Id())
}

func TestResourceCloudAccountCloudStackRead(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountCloudStackRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cloudstack-account-id-1", d.Id())
}

func TestResourceCloudAccountCloudStackUpdate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountCloudStackUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-cloudstack-account-id-1", d.Id())
}

func TestResourceCloudAccountCloudStackDelete(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountCloudStackDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}
