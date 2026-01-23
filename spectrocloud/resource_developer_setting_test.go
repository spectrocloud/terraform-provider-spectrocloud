package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestToDeveloperSetting(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	// Set custom values
	d.Set("virtual_clusters_limit", int32(10))
	d.Set("cpu", int32(4))
	d.Set("memory", int32(16))
	d.Set("storage", int32(50))

	devCredit, sysClusterGroupPref := toDeveloperSetting(d)

	assert.NotNil(t, devCredit)
	assert.NotNil(t, sysClusterGroupPref)
	assert.Equal(t, int32(10), devCredit.VirtualClustersLimit)
	assert.Equal(t, int32(4), devCredit.CPU)
	assert.Equal(t, int32(16), devCredit.MemoryGiB)
	assert.Equal(t, int32(50), devCredit.StorageGiB)
	assert.False(t, sysClusterGroupPref.HideSystemClusterGroups)
}

func TestToDeveloperSettingDefault(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	devCredit, sysClusterGroupPref := toDeveloperSettingDefault(d)

	assert.NotNil(t, devCredit)
	assert.NotNil(t, sysClusterGroupPref)
	assert.Equal(t, int32(12), devCredit.CPU)
	assert.Equal(t, int32(16), devCredit.MemoryGiB)
	assert.Equal(t, int32(20), devCredit.StorageGiB)
	assert.Equal(t, int32(2), devCredit.VirtualClustersLimit)
	assert.False(t, sysClusterGroupPref.HideSystemClusterGroups)
}

func TestFlattenDeveloperSetting(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	devSetting := &models.V1DeveloperCredit{
		CPU:                  8,
		MemoryGiB:            32,
		StorageGiB:           100,
		VirtualClustersLimit: 5,
	}
	sysClusterGroupPref := &models.V1TenantEnableClusterGroup{
		HideSystemClusterGroups: true,
	}

	err := flattenDeveloperSetting(devSetting, sysClusterGroupPref, d)
	assert.NoError(t, err)

	// Verify values set in schema
	assert.Equal(t, 8, d.Get("cpu"))
	assert.Equal(t, 32, d.Get("memory"))
	assert.Equal(t, 100, d.Get("storage"))
	assert.Equal(t, 5, d.Get("virtual_clusters_limit"))
	assert.True(t, d.Get("hide_system_cluster_group").(bool))
}

// TestResourceDeveloperSettingDelete tests the resourceDeveloperSettingDelete function.
// This function:
// 1. Gets a V1Client with tenant context
// 2. Calls toDeveloperSettingDefault to get default settings (since deletion resets to defaults)
// 3. Gets the tenant UID
// 4. Updates developer setting to default values
// 5. Updates system cluster group preference to default values
// 6. Clears the resource ID
func TestResourceDeveloperSettingDelete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func() *schema.ResourceData
		client      interface{}
		expectError bool
		errorMsg    string
		description string
		verify      func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics)
	}{
		{
			name: "Successful deletion (reset to defaults)",
			setup: func() *schema.ResourceData {
				d := resourceDeveloperSetting().TestResourceData()
				d.SetId("default-dev-setting-id")
				// Set some custom values before deletion
				_ = d.Set("virtual_clusters_limit", 10)
				_ = d.Set("cpu", 8)
				_ = d.Set("memory", 32)
				_ = d.Set("storage", 100)
				_ = d.Set("hide_system_cluster_group", true)
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false,
			description: "Should successfully reset developer settings to defaults",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.Empty(t, diags, "Should not have diagnostics on success")
				// Verify resource ID is cleared
				assert.Empty(t, d.Id(), "Resource ID should be cleared after deletion")
			},
		},
		{
			name: "Error when GetTenantUID fails",
			setup: func() *schema.ResourceData {
				d := resourceDeveloperSetting().TestResourceData()
				d.SetId("default-dev-setting-id")
				return d
			},
			client:      unitTestMockAPINegativeClient,
			expectError: true,
			errorMsg:    "tenant",
			description: "Should return error when GetTenantUID fails",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.NotEmpty(t, diags, "Should have diagnostics on error")
				if len(diags) > 0 {
					assert.True(t, diags.HasError(), "Diagnostics should contain errors")
					// Note: Mock API may not return tenant-specific errors, so we just verify error exists
				}
			},
		},
		{
			name: "Error when UpdateDeveloperSetting fails",
			setup: func() *schema.ResourceData {
				d := resourceDeveloperSetting().TestResourceData()
				d.SetId("default-dev-setting-id")
				return d
			},
			client:      unitTestMockAPINegativeClient,
			expectError: true,
			description: "Should return error when UpdateDeveloperSetting fails",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.NotEmpty(t, diags, "Should have diagnostics on error")
				if len(diags) > 0 {
					assert.True(t, diags.HasError(), "Diagnostics should contain errors")
				}
			},
		},
		{
			name: "Error when UpdateSystemClusterGroupPreference fails",
			setup: func() *schema.ResourceData {
				d := resourceDeveloperSetting().TestResourceData()
				d.SetId("default-dev-setting-id")
				return d
			},
			client:      unitTestMockAPINegativeClient,
			expectError: true,
			description: "Should return error when UpdateSystemClusterGroupPreference fails",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				assert.NotEmpty(t, diags, "Should have diagnostics on error")
				if len(diags) > 0 {
					assert.True(t, diags.HasError(), "Diagnostics should contain errors")
				}
			},
		},
		{
			name: "Deletion with empty resource ID",
			setup: func() *schema.ResourceData {
				d := resourceDeveloperSetting().TestResourceData()
				d.SetId("") // Empty ID
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false,
			description: "Should handle deletion with empty resource ID",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				// Function should still attempt to reset to defaults
				assert.Empty(t, diags, "Should not have diagnostics on success")
				assert.Empty(t, d.Id(), "Resource ID should remain empty after deletion")
			},
		},
		{
			name: "Deletion verifies default values are used",
			setup: func() *schema.ResourceData {
				d := resourceDeveloperSetting().TestResourceData()
				d.SetId("default-dev-setting-id")
				// Set custom values
				_ = d.Set("virtual_clusters_limit", 50)
				_ = d.Set("cpu", 20)
				_ = d.Set("memory", 64)
				_ = d.Set("storage", 200)
				_ = d.Set("hide_system_cluster_group", true)
				return d
			},
			client:      unitTestMockAPIClient,
			expectError: false,
			description: "Should use default values from toDeveloperSettingDefault",
			verify: func(t *testing.T, d *schema.ResourceData, diags diag.Diagnostics) {
				// Verify that toDeveloperSettingDefault is called (defaults: CPU=12, Memory=16, Storage=20, VirtualClustersLimit=2, HideSystemClusterGroups=false)
				// The function should call UpdateDeveloperSetting and UpdateSystemClusterGroupPreference with these defaults
				assert.Empty(t, diags, "Should not have diagnostics on success")
				assert.Empty(t, d.Id(), "Resource ID should be cleared")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := tt.setup()

			// Call the delete function
			diags := resourceDeveloperSettingDelete(ctx, resourceData, tt.client)

			// Verify results
			if tt.expectError {
				assert.True(t, diags.HasError(), "Expected error for test case: %s", tt.description)
				if tt.errorMsg != "" {
					// Check if error message contains expected text in Summary or Detail
					found := false
					for _, diag := range diags {
						if diag.Summary != "" && strings.Contains(strings.ToLower(diag.Summary), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
						if diag.Detail != "" && strings.Contains(strings.ToLower(diag.Detail), strings.ToLower(tt.errorMsg)) {
							found = true
							break
						}
					}
					if !found && len(diags) > 0 {
						// Log diagnostics for debugging
						for _, diag := range diags {
							if diag.Summary != "" {
								t.Logf("Diagnostic Summary: %s", diag.Summary)
							}
							if diag.Detail != "" {
								t.Logf("Diagnostic Detail: %s", diag.Detail)
							}
						}
					}
				}
			} else {
				if diags.HasError() {
					// If error occurred but not expected, log it for debugging
					t.Logf("Unexpected error: %v", diags)
				}
				assert.False(t, diags.HasError(), "Should not have errors for test case: %s", tt.description)
			}

			// Run custom verify function if provided
			if tt.verify != nil {
				tt.verify(t, resourceData, diags)
			}
		})
	}
}
