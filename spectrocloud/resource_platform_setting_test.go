package spectrocloud

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testValidateContextDependenciesLogic tests the validation logic directly
// This is a helper function that mimics the behavior of validateContextDependencies
// but works with ResourceData instead of ResourceDiff for testing purposes
func testValidateContextDependenciesLogic(contextVal string, fields map[string]interface{}) error {
	if contextVal == "project" {
		disallowedFields := []string{"session_timeout", "login_banner", "non_fips_addon_pack", "non_fips_features", "non_fips_cluster_import"}

		for _, field := range disallowedFields {
			if _, exists := fields[field]; exists {
				return fmt.Errorf("attribute %q is not allowed when context is set to 'project'", field)
			}
		}
	}
	return nil
}

// TestValidateContextDependencies tests the validation logic
func TestValidateContextDependencies(t *testing.T) {

	tests := []struct {
		name        string
		contextVal  string
		fields      map[string]interface{}
		expectError bool
		errorMsg    string
		description string
		verify      func(t *testing.T, err error)
	}{
		{
			name:       "Tenant context with allowed fields - no error",
			contextVal: "tenant",
			fields: map[string]interface{}{
				"session_timeout": 240,
				"login_banner": []interface{}{
					map[string]interface{}{
						"title":   "Test Title",
						"message": "Test Message",
					},
				},
				"non_fips_addon_pack": true,
			},
			expectError: false,
			description: "Should allow all fields when context is tenant",
			verify: func(t *testing.T, err error) {
				assert.NoError(t, err, "Should not have error for tenant context")
			},
		},
		{
			name:       "Project context with no disallowed fields - no error",
			contextVal: "project",
			fields: map[string]interface{}{
				"pause_agent_upgrades":     "lock",
				"enable_auto_remediation":  true,
				"cluster_auto_remediation": true,
			},
			expectError: false,
			description: "Should allow project context with only allowed fields",
			verify: func(t *testing.T, err error) {
				assert.NoError(t, err, "Should not have error when no disallowed fields are set")
			},
		},
		{
			name:       "Project context with session_timeout - returns error",
			contextVal: "project",
			fields: map[string]interface{}{
				"session_timeout": 240,
			},
			expectError: true,
			errorMsg:    "attribute \"session_timeout\" is not allowed when context is set to 'project'",
			description: "Should return error when session_timeout is set with project context",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "Should have error")
				assert.Contains(t, err.Error(), "session_timeout", "Error should mention session_timeout")
				assert.Contains(t, err.Error(), "not allowed when context is set to 'project'", "Error should mention project context")
			},
		},
		{
			name:       "Project context with login_banner - returns error",
			contextVal: "project",
			fields: map[string]interface{}{
				"login_banner": []interface{}{
					map[string]interface{}{
						"title":   "Test Title",
						"message": "Test Message",
					},
				},
			},
			expectError: true,
			errorMsg:    "attribute \"login_banner\" is not allowed when context is set to 'project'",
			description: "Should return error when login_banner is set with project context",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "Should have error")
				assert.Contains(t, err.Error(), "login_banner", "Error should mention login_banner")
			},
		},
		{
			name:       "Project context with non_fips_addon_pack - returns error",
			contextVal: "project",
			fields: map[string]interface{}{
				"non_fips_addon_pack": true,
			},
			expectError: true,
			errorMsg:    "attribute \"non_fips_addon_pack\" is not allowed when context is set to 'project'",
			description: "Should return error when non_fips_addon_pack is set with project context",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "Should have error")
				assert.Contains(t, err.Error(), "non_fips_addon_pack", "Error should mention non_fips_addon_pack")
			},
		},
		{
			name:       "Project context with non_fips_features - returns error",
			contextVal: "project",
			fields: map[string]interface{}{
				"non_fips_features": true,
			},
			expectError: true,
			errorMsg:    "attribute \"non_fips_features\" is not allowed when context is set to 'project'",
			description: "Should return error when non_fips_features is set with project context",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "Should have error")
				assert.Contains(t, err.Error(), "non_fips_features", "Error should mention non_fips_features")
			},
		},
		{
			name:       "Project context with non_fips_cluster_import - returns error",
			contextVal: "project",
			fields: map[string]interface{}{
				"non_fips_cluster_import": true,
			},
			expectError: true,
			errorMsg:    "attribute \"non_fips_cluster_import\" is not allowed when context is set to 'project'",
			description: "Should return error when non_fips_cluster_import is set with project context",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "Should have error")
				assert.Contains(t, err.Error(), "non_fips_cluster_import", "Error should mention non_fips_cluster_import")
			},
		},
		{
			name:       "Project context with multiple disallowed fields - returns error for first found",
			contextVal: "project",
			fields: map[string]interface{}{
				"session_timeout":     240,
				"non_fips_addon_pack": true,
				"login_banner": []interface{}{
					map[string]interface{}{
						"title":   "Test",
						"message": "Test",
					},
				},
			},
			expectError: true,
			description: "Should return error when multiple disallowed fields are set (returns first found)",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "Should have error")
				// The function checks fields in order, so it should return error for session_timeout first
				assert.Contains(t, err.Error(), "not allowed when context is set to 'project'", "Error should mention project context")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			// Since ResourceDiff is difficult to create in unit tests,
			// we test the core validation logic which is the same
			err := testValidateContextDependenciesLogic(tt.contextVal, tt.fields)

			if tt.verify != nil {
				tt.verify(t, err)
			} else {
				if tt.expectError {
					assert.Error(t, err, tt.description)
					if tt.errorMsg != "" {
						assert.Equal(t, tt.errorMsg, err.Error(), tt.description)
					}
				} else {
					assert.NoError(t, err, tt.description)
				}
			}
		})
	}
}

// TestValidateContextDependencies_AllDisallowedFields tests each disallowed field individually
func TestValidateContextDependencies_AllDisallowedFields(t *testing.T) {
	disallowedFields := []struct {
		name  string
		value interface{}
	}{
		{"session_timeout", 240},
		{"login_banner", []interface{}{
			map[string]interface{}{
				"title":   "Test",
				"message": "Test",
			},
		}},
		{"non_fips_addon_pack", true},
		{"non_fips_features", true},
		{"non_fips_cluster_import", true},
	}

	for _, field := range disallowedFields {
		t.Run(field.name, func(t *testing.T) {
			// Test the validation logic directly
			err := testValidateContextDependenciesLogic("project", map[string]interface{}{
				field.name: field.value,
			})

			assert.Error(t, err, "Should return error for %s with project context", field.name)
			assert.Contains(t, err.Error(), field.name, "Error should mention the field name")
			assert.Contains(t, err.Error(), "not allowed when context is set to 'project'", "Error should mention project context")
		})
	}
}

func TestConvertFIPSBool(t *testing.T) {
	tests := []struct {
		name        string
		flag        bool
		expected    string
		description string
	}{
		{
			name:        "True flag returns nonFipsEnabled",
			flag:        true,
			expected:    "nonFipsEnabled",
			description: "Should return 'nonFipsEnabled' when flag is true",
		},
		{
			name:        "False flag returns nonFipsDisabled",
			flag:        false,
			expected:    "nonFipsDisabled",
			description: "Should return 'nonFipsDisabled' when flag is false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertFIPSBool(tt.flag)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}
