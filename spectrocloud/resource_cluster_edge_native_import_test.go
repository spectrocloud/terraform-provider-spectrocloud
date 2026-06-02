package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResourceID(t *testing.T) {
	tests := []struct {
		name          string
		importID      string
		expectedScope string
		expectedID    string
		expectError   bool
	}{
		{
			name:          "valid project scope",
			importID:      "cluster-123:project",
			expectedScope: "project",
			expectedID:    "cluster-123",
		},
		{
			name:          "valid tenant scope",
			importID:      "cluster-456:tenant",
			expectedScope: "tenant",
			expectedID:    "cluster-456",
		},
		{
			name:        "missing scope delimiter",
			importID:    "cluster-123",
			expectError: true,
		},
		{
			name:        "unsupported scope",
			importID:    "cluster-123:workspace",
			expectError: true,
		},
		{
			name:        "too many id segments",
			importID:    "cluster-123:project:extra",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{})
			d.SetId(tt.importID)

			scope, clusterID, err := ParseResourceID(d)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid cluster ID format specified for import")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedScope, scope)
			assert.Equal(t, tt.expectedID, clusterID)
		})
	}
}
