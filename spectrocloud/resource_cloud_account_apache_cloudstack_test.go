package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
)

// Test for the `toApacheCloudStackAccount` function
func TestToApacheCloudStackAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1CloudStackAccount
	}{
		{
			name: "All Fields Present",
			input: map[string]interface{}{
				"name":                     "apache-cloudstack-account",
				"private_cloud_gateway_id": "private-cloud-gateway-id",
				"api_url":                  "https://cloudstack.example.com:8080/client/api",
				"api_key":                  "api-key",
				"secret_key":               "secret-key",
				"domain":                   "ROOT",
				"insecure":                 false,
			},
			expected: &models.V1CloudStackAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "apache-cloudstack-account",
					Annotations: map[string]string{OverlordUID: "private-cloud-gateway-id"},
					UID:         "",
				},
				Spec: &models.V1CloudStackCloudAccount{
					APIURL:    types.Ptr("https://cloudstack.example.com:8080/client/api"),
					APIKey:    types.Ptr("api-key"),
					SecretKey: types.Ptr("secret-key"),
					Domain:    "ROOT",
					Insecure:  false,
				},
			},
		},
		{
			name: "Insecure Mode Enabled",
			input: map[string]interface{}{
				"name":                     "apache-cloudstack-insecure",
				"private_cloud_gateway_id": "pcg-id",
				"api_url":                  "https://cloudstack.example.com:8080/client/api",
				"api_key":                  "test-key",
				"secret_key":               "test-secret",
				"domain":                   "",
				"insecure":                 true,
			},
			expected: &models.V1CloudStackAccount{
				Metadata: &models.V1ObjectMeta{
					Name:        "apache-cloudstack-insecure",
					Annotations: map[string]string{OverlordUID: "pcg-id"},
					UID:         "",
				},
				Spec: &models.V1CloudStackCloudAccount{
					APIURL:    types.Ptr("https://cloudstack.example.com:8080/client/api"),
					APIKey:    types.Ptr("test-key"),
					SecretKey: types.Ptr("test-secret"),
					Domain:    "",
					Insecure:  true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceCloudAccountApacheCloudStack().Schema, tt.input)

			// Call the function under test (passing nil client since we're only testing conversion logic)
			c := unitTestMockAPIClient.(*client.V1Client)
			result := toApacheCloudStackAccount(d, c)

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

			assert.Equal(t, tt.expected.Spec.Domain, result.Spec.Domain)
			assert.Equal(t, tt.expected.Spec.Insecure, result.Spec.Insecure)
		})
	}
}

func prepareResourceCloudAccountApacheCloudStack() *schema.ResourceData {
	d := resourceCloudAccountApacheCloudStack().TestResourceData()
	d.SetId("test-apache-cloudstack-account-id-1")
	_ = d.Set("name", "test-apache-cloudstack-account-1")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "test-pcg-id")
	_ = d.Set("api_url", "https://test.cloudstack.com:8080/client/api")
	_ = d.Set("api_key", "test-api-key")
	_ = d.Set("secret_key", "test-secret-key")
	_ = d.Set("domain", "ROOT")
	_ = d.Set("insecure", false)
	return d
}

func TestResourceCloudAccountApacheCloudStackCreate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountApacheCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountApacheCloudStackCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-apache-cloudstack-account-id-1", d.Id())
}

func TestResourceCloudAccountApacheCloudStackRead(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountApacheCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountApacheCloudStackRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-apache-cloudstack-account-id-1", d.Id())
}

func TestResourceCloudAccountApacheCloudStackUpdate(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountApacheCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountApacheCloudStackUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-apache-cloudstack-account-id-1", d.Id())
}

func TestResourceCloudAccountApacheCloudStackDelete(t *testing.T) {
	// Mock context and resource data
	d := prepareResourceCloudAccountApacheCloudStack()
	ctx := context.Background()
	diags := resourceCloudAccountApacheCloudStackDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
}

// Test for PCG Type Detection - System Private Gateway
func TestToApacheCloudStackAccountWithSystemPCG(t *testing.T) {
	// Create resource data with System Private Gateway
	d := resourceCloudAccountApacheCloudStack().TestResourceData()
	_ = d.Set("name", "test-cloudstack-account-system-pcg")
	_ = d.Set("private_cloud_gateway_id", "test-system-pcg-id")
	_ = d.Set("api_url", "https://cloudstack.example.com:8080/client/api")
	_ = d.Set("api_key", "test-api-key")
	_ = d.Set("secret_key", "test-secret-key")
	_ = d.Set("domain", "ROOT")
	_ = d.Set("insecure", false)

	// Call the function under test with mock client
	c := unitTestMockAPIClient.(*client.V1Client)
	account := toApacheCloudStackAccount(d, c)

	// Assert that overlordType annotation is set to "system" for System Private Gateway
	assert.Equal(t, "test-cloudstack-account-system-pcg", account.Metadata.Name)
	assert.Equal(t, "test-system-pcg-id", account.Metadata.Annotations[OverlordUID])
	assert.Equal(t, "system", account.Metadata.Annotations["overlordType"], "overlordType should be 'system' for System Private Gateway")
}

// Test for PCG Type Detection - Regular/Custom PCG
func TestToApacheCloudStackAccountWithRegularPCG(t *testing.T) {
	// Create resource data with regular PCG
	d := resourceCloudAccountApacheCloudStack().TestResourceData()
	_ = d.Set("name", "test-cloudstack-account-regular-pcg")
	_ = d.Set("private_cloud_gateway_id", "test-regular-pcg-id")
	_ = d.Set("api_url", "https://cloudstack.example.com:8080/client/api")
	_ = d.Set("api_key", "test-api-key")
	_ = d.Set("secret_key", "test-secret-key")
	_ = d.Set("domain", "ROOT")
	_ = d.Set("insecure", false)

	// Call the function under test with mock client
	c := unitTestMockAPIClient.(*client.V1Client)
	account := toApacheCloudStackAccount(d, c)

	// Assert that overlordType annotation is NOT set for regular PCG
	assert.Equal(t, "test-cloudstack-account-regular-pcg", account.Metadata.Name)
	assert.Equal(t, "test-regular-pcg-id", account.Metadata.Annotations[OverlordUID])
	assert.NotContains(t, account.Metadata.Annotations, "overlordType", "overlordType should NOT be set for regular PCG")
}

// Test for Create with System PCG
func TestResourceCloudAccountApacheCloudStackCreateWithSystemPCG(t *testing.T) {
	// Create resource data with System Private Gateway
	d := resourceCloudAccountApacheCloudStack().TestResourceData()
	_ = d.Set("name", "test-cloudstack-account-system")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "test-system-pcg-id")
	_ = d.Set("api_url", "https://cloudstack.example.com:8080/client/api")
	_ = d.Set("api_key", "test-api-key")
	_ = d.Set("secret_key", "test-secret-key")
	_ = d.Set("domain", "ROOT")
	_ = d.Set("insecure", false)

	ctx := context.Background()
	diags := resourceCloudAccountApacheCloudStackCreate(ctx, d, unitTestMockAPIClient)

	// Assert no errors
	assert.Len(t, diags, 0)
	assert.NotEmpty(t, d.Id(), "Resource ID should be set after creation")
}

// Test for Create with Regular PCG
func TestResourceCloudAccountApacheCloudStackCreateWithRegularPCG(t *testing.T) {
	// Create resource data with regular PCG
	d := resourceCloudAccountApacheCloudStack().TestResourceData()
	_ = d.Set("name", "test-cloudstack-account-regular")
	_ = d.Set("context", "project")
	_ = d.Set("private_cloud_gateway_id", "test-regular-pcg-id")
	_ = d.Set("api_url", "https://cloudstack.example.com:8080/client/api")
	_ = d.Set("api_key", "test-api-key")
	_ = d.Set("secret_key", "test-secret-key")
	_ = d.Set("domain", "Production")
	_ = d.Set("insecure", true)

	ctx := context.Background()
	diags := resourceCloudAccountApacheCloudStackCreate(ctx, d, unitTestMockAPIClient)

	// Assert no errors
	assert.Len(t, diags, 0)
	assert.NotEmpty(t, d.Id(), "Resource ID should be set after creation")
}
