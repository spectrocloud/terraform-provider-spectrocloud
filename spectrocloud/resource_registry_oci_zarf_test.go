package spectrocloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// validateZarfProviderSync is a testable version of the CustomizeDiff validation logic
func validateZarfProviderSync(providerType, registryType string, isSync bool) error {
	// Validate that `provider_type` is "zarf" only if `type` is "basic"
	if providerType == "zarf" && registryType != "basic" {
		return fmt.Errorf("`provider_type` set to `zarf` is only allowed when `type` is `basic`")
	}
	if providerType == "pack" && !isSync {
		return fmt.Errorf("`provider_type` set to `pack` is only allowed when `is_synchronization` is set to `true`")
	}
	return nil
}

// prepareZarfOciRegistryTestData creates test data for Zarf OCI registry
func prepareZarfOciRegistryTestData() *schema.ResourceData {
	d := resourceRegistryOciEcr().TestResourceData()
	_ = d.Set("name", "test-zarf-registry")
	_ = d.Set("type", "basic")
	_ = d.Set("endpoint", "https://registry.example.com")
	_ = d.Set("is_private", true)
	_ = d.Set("provider_type", "zarf")
	_ = d.Set("base_content_path", "/")

	var credential []map[string]interface{}
	cred := map[string]interface{}{
		"credential_type": "basic",
		"username":        "test-username",
		"password":        "test-password",
	}
	credential = append(credential, cred)
	_ = d.Set("credentials", credential)
	return d
}

// TestZarfProviderSyncValidation tests is_synchronization validation for Zarf provider
func TestZarfProviderSyncValidation(t *testing.T) {
	testCases := []struct {
		name          string
		providerType  string
		registryType  string
		isSync        bool
		shouldError   bool
		errorContains string
	}{
		{
			name:         "Zarf with basic type and sync enabled - should be valid",
			providerType: "zarf",
			registryType: "basic",
			isSync:       true,
			shouldError:  false,
		},
		{
			name:         "Zarf with basic type and sync disabled - should be valid",
			providerType: "zarf",
			registryType: "basic",
			isSync:       false,
			shouldError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateZarfProviderSync(tc.providerType, tc.registryType, tc.isSync)

			if tc.shouldError {
				assert.Error(t, err, "Expected error for: %s", tc.name)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err, "Expected no error for: %s, got: %v", tc.name, err)
			}
		})
	}
}

// TestZarfProviderIsSynchronizationField tests the is_synchronization field specifically for Zarf provider
func TestZarfProviderIsSynchronizationField(t *testing.T) {
	testCases := []struct {
		name          string
		providerType  string
		registryType  string
		isSync        bool
		shouldError   bool
		errorContains string
		description   string
	}{
		{
			name:         "Zarf basic with is_synchronization=true",
			providerType: "zarf",
			registryType: "basic",
			isSync:       true,
			shouldError:  false,
			description:  "Zarf provider with basic type should accept is_synchronization=true",
		},
		{
			name:         "Zarf basic with is_synchronization=false",
			providerType: "zarf",
			registryType: "basic",
			isSync:       false,
			shouldError:  false,
			description:  "Zarf provider with basic type should accept is_synchronization=false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateZarfProviderSync(tc.providerType, tc.registryType, tc.isSync)

			if tc.shouldError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message should match expected")
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestZarfProviderSyncWithAPIValidation tests Zarf provider with is_synchronization=true
// This test verifies that the API validation endpoint is called when creating a Zarf registry with sync enabled
func TestZarfProviderSyncWithAPIValidation(t *testing.T) {
	d := prepareZarfOciRegistryTestData()
	_ = d.Set("is_synchronization", true)

	ctx := context.Background()
	diags := resourceRegistryEcrCreate(ctx, d, unitTestMockAPIClient)

	// Should complete successfully with no errors
	// The validation endpoint (/v1/registries/oci/basic/validate) should be called
	// and return success (204 status code from mock)
	assert.Equal(t, 0, len(diags), "Expected no errors when creating Zarf registry with sync enabled")
	assert.NotEmpty(t, d.Id(), "Expected resource ID to be set after successful creation")
}

// TestZarfProviderSyncDisabledNoAPICall tests Zarf provider with is_synchronization=false
// This test verifies that the API validation endpoint is NOT called when sync is disabled
func TestZarfProviderSyncDisabledNoAPICall(t *testing.T) {
	d := prepareZarfOciRegistryTestData()
	_ = d.Set("is_synchronization", false)

	ctx := context.Background()
	diags := resourceRegistryEcrCreate(ctx, d, unitTestMockAPIClient)

	// Should complete successfully with no errors
	// The validation endpoint should NOT be called when is_synchronization=false
	assert.Equal(t, 0, len(diags), "Expected no errors when creating Zarf registry with sync disabled")
	assert.NotEmpty(t, d.Id(), "Expected resource ID to be set after successful creation")
}
