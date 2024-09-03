package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceBackupStorageLocationRead(t *testing.T) {
	tests := []struct {
		name           string
		inputID        string
		inputName      string
		bsls           []*models.V1UserAssetsLocation
		expectedDiag   diag.Diagnostics
		expectedID     string
		expectedName   string
		expectingError bool
	}{
		{
			name:    "Backup storage location not found",
			inputID: "non-existent-uid",
			bsls:    []*models.V1UserAssetsLocation{{Metadata: &models.V1ObjectMeta{UID: "test-bsl-location-uid", Name: "test-bsl-location"}}},
			expectedDiag: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Unable to find backup storage location",
					Detail:   "Unable to find the specified backup storage location",
				},
			},
			expectingError: true,
		},
		{
			name:    "Error setting name in state",
			inputID: "test-bsl-location-uid",
			bsls:    []*models.V1UserAssetsLocation{{Metadata: &models.V1ObjectMeta{UID: "test-bsl-location-uid", Name: "test-bsl-location"}}},
			expectedDiag: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Unable to find backup storage location",
					Detail:   "Unable to find the specified backup storage location",
				},
			},
			expectingError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := dataSourceBackupStorageLocation().TestResourceData()

			resourceData.SetId(tt.inputID)
			if tt.inputName != "" {
				resourceData.Set("name", tt.inputName)
			}

			diags := dataSourceBackupStorageLocationRead(context.Background(), resourceData, unitTestMockAPIClient)

			if tt.expectingError {
				assert.Equal(t, tt.expectedDiag, diags)
			} else {
				assert.Equal(t, "", diags)
				assert.Equal(t, tt.expectedID, resourceData.Id())
				name, _ := resourceData.Get("name").(string)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}
