package spectrocloud

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClusterProfileImportID(t *testing.T) {
	tests := []struct {
		name            string
		importID        string
		expectedID      string
		expectedContext string
		expectedVersion string
		expectError     bool
	}{
		{
			name:            "uid and context",
			importID:        "profile-uid:project",
			expectedID:      "profile-uid",
			expectedContext: "project",
			expectedVersion: "",
		},
		{
			name:            "name context and version",
			importID:        "base-profile:tenant:2.3.4",
			expectedID:      "base-profile",
			expectedContext: "tenant",
			expectedVersion: "2.3.4",
		},
		{
			name:        "invalid segment count",
			importID:    "a:b:c:d",
			expectError: true,
		},
		{
			name:        "empty id",
			importID:    ":project",
			expectError: true,
		},
		{
			name:        "invalid context",
			importID:    "abc:workspace",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ctx, version, err := ParseClusterProfileImportID(tt.importID)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedContext, ctx)
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseBackupStorageLocationImportID(t *testing.T) {
	tests := []struct {
		name          string
		importID      string
		expectedScope string
		expectedID    string
		expectError   bool
	}{
		{
			name:          "id only defaults to project",
			importID:      "bsl-uid",
			expectedScope: "project",
			expectedID:    "bsl-uid",
		},
		{
			name:          "id and tenant context",
			importID:      "bsl-name:tenant",
			expectedScope: "tenant",
			expectedID:    "bsl-name",
		},
		{
			name:        "empty id",
			importID:    "",
			expectError: true,
		},
		{
			name:        "invalid context",
			importID:    "bsl:system",
			expectError: true,
		},
		{
			name:        "too many segments",
			importID:    "a:b:c",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{})
			d.SetId(tt.importID)
			scope, id, err := parseBackupStorageLocationImportID(d)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedScope, scope)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestMapAPITypeToTerraformProvider(t *testing.T) {
	assert.Equal(t, "aws", mapAPITypeToTerraformProvider("s3"))
	assert.Equal(t, "gcp", mapAPITypeToTerraformProvider("gcp"))
	assert.Equal(t, "minio", mapAPITypeToTerraformProvider("minio"))
	assert.Equal(t, "azure", mapAPITypeToTerraformProvider("azure"))
	assert.Equal(t, "aws", mapAPITypeToTerraformProvider("unknown"))
}

func TestIsUserNotFound(t *testing.T) {
	assert.False(t, isUserNotFound(nil))
	assert.False(t, isUserNotFound(errors.New("boom")))
	assert.True(t, isUserNotFound(errors.New("Code:UserNotFound user is missing")))
	assert.True(t, isUserNotFound(errors.New("Specified user not found for tenant")))
}
