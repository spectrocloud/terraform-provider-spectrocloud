package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestGetUpdateStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "Default when no strategy specified",
			input:    map[string]interface{}{},
			expected: "RollingUpdateScaleOut",
		},
		{
			name: "Legacy update_strategy",
			input: map[string]interface{}{
				"update_strategy": "RollingUpdateScaleIn",
			},
			expected: "RollingUpdateScaleIn",
		},
		{
			name: "New rolling_update_strategy takes precedence",
			input: map[string]interface{}{
				"update_strategy": "RollingUpdateScaleIn",
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type": "OverrideScaling",
					},
				},
			},
			expected: "OverrideScaling",
		},
		{
			name: "New rolling_update_strategy with all fields",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type":            "OverrideScaling",
						"max_surge":       "1",
						"max_unavailable": "0",
					},
				},
			},
			expected: "OverrideScaling",
		},
		{
			name: "Rolling update strategy with empty list",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{},
				"update_strategy":         "RollingUpdateScaleIn",
			},
			expected: "RollingUpdateScaleIn",
		},
		{
			name: "Rolling update strategy with invalid type (not a map) - should fallback",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{"invalid"},
				"update_strategy":         "RollingUpdateScaleIn",
			},
			expected: "RollingUpdateScaleIn",
		},
		{
			name: "Rolling update strategy with nil type field - should fallback",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type": nil,
					},
				},
				"update_strategy": "RollingUpdateScaleIn",
			},
			expected: "RollingUpdateScaleIn",
		},
		{
			name: "Rolling update strategy with non-string type - should default",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type": 123,
					},
				},
			},
			expected: "RollingUpdateScaleOut",
		},
		{
			name: "Update strategy with nil value - should default",
			input: map[string]interface{}{
				"update_strategy": nil,
			},
			expected: "RollingUpdateScaleOut",
		},
		{
			name: "Update strategy with non-string value - should default",
			input: map[string]interface{}{
				"update_strategy": 123,
			},
			expected: "RollingUpdateScaleOut",
		},
		{
			name: "Update strategy with empty string - should default",
			input: map[string]interface{}{
				"update_strategy": "",
			},
			expected: "RollingUpdateScaleOut",
		},
		{
			name: "Rolling update strategy with empty type string - should fallback",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type": "",
					},
				},
				"update_strategy": "RollingUpdateScaleIn",
			},
			expected: "RollingUpdateScaleIn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUpdateStrategy(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToUpdateStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1UpdateStrategy
	}{
		{
			name:  "Default strategy",
			input: map[string]interface{}{},
			expected: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
		},
		{
			name: "Legacy update_strategy only",
			input: map[string]interface{}{
				"update_strategy": "RollingUpdateScaleIn",
			},
			expected: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleIn",
			},
		},
		{
			name: "New rolling_update_strategy with type only",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type": "OverrideScaling",
					},
				},
			},
			expected: &models.V1UpdateStrategy{
				Type: "OverrideScaling",
			},
		},
		{
			name: "New rolling_update_strategy with all fields",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type":            "OverrideScaling",
						"max_surge":       "1",
						"max_unavailable": "0",
					},
				},
			},
			expected: &models.V1UpdateStrategy{
				Type:           "OverrideScaling",
				MaxSurge:       "1",
				MaxUnavailable: "0",
			},
		},
		{
			name: "Rolling update strategy with percentage values",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type":            "OverrideScaling",
						"max_surge":       "20%",
						"max_unavailable": "10%",
					},
				},
			},
			expected: &models.V1UpdateStrategy{
				Type:           "OverrideScaling",
				MaxSurge:       "20%",
				MaxUnavailable: "10%",
			},
		},
		{
			name: "Rolling update strategy ignores empty max_surge/max_unavailable",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type":            "RollingUpdateScaleOut",
						"max_surge":       "",
						"max_unavailable": "",
					},
				},
			},
			expected: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
		},
		{
			name: "Rolling update strategy with invalid type (not a map) - should get default type",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{"invalid"},
			},
			expected: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
		},
		{
			name: "Rolling update strategy with nil max_surge/max_unavailable",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type":            "OverrideScaling",
						"max_surge":       nil,
						"max_unavailable": nil,
					},
				},
			},
			expected: &models.V1UpdateStrategy{
				Type: "OverrideScaling",
			},
		},
		{
			name: "Rolling update strategy with non-string max values - should ignore them",
			input: map[string]interface{}{
				"rolling_update_strategy": []interface{}{
					map[string]interface{}{
						"type":            "OverrideScaling",
						"max_surge":       123,
						"max_unavailable": 456,
					},
				},
			},
			expected: &models.V1UpdateStrategy{
				Type: "OverrideScaling",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toUpdateStrategy(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestFlattenUpdateStrategy(t *testing.T) {
	tests := []struct {
		name            string
		input           *models.V1UpdateStrategy
		existingFields  map[string]interface{}
		expectedLegacy  string
		expectedRolling []interface{}
	}{
		{
			name:            "Nil strategy defaults to RollingUpdateScaleOut",
			input:           nil,
			existingFields:  map[string]interface{}{},
			expectedLegacy:  "RollingUpdateScaleOut",
			expectedRolling: nil,
		},
		{
			name: "Empty strategy type defaults to RollingUpdateScaleOut",
			input: &models.V1UpdateStrategy{
				Type: "",
			},
			existingFields:  map[string]interface{}{},
			expectedLegacy:  "RollingUpdateScaleOut",
			expectedRolling: nil,
		},
		{
			name: "Simple strategy type without maxSurge/maxUnavailable",
			input: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleIn",
			},
			existingFields:  map[string]interface{}{},
			expectedLegacy:  "RollingUpdateScaleIn",
			expectedRolling: nil,
		},
		{
			name: "Strategy with maxSurge and maxUnavailable sets rolling_update_strategy",
			input: &models.V1UpdateStrategy{
				Type:           "OverrideScaling",
				MaxSurge:       "1",
				MaxUnavailable: "0",
			},
			existingFields: map[string]interface{}{},
			expectedLegacy: "OverrideScaling",
			expectedRolling: []interface{}{
				map[string]interface{}{
					"type":            "OverrideScaling",
					"max_surge":       "1",
					"max_unavailable": "0",
				},
			},
		},
		{
			name: "Strategy with percentage values",
			input: &models.V1UpdateStrategy{
				Type:           "OverrideScaling",
				MaxSurge:       "20%",
				MaxUnavailable: "10%",
			},
			existingFields: map[string]interface{}{},
			expectedLegacy: "OverrideScaling",
			expectedRolling: []interface{}{
				map[string]interface{}{
					"type":            "OverrideScaling",
					"max_surge":       "20%",
					"max_unavailable": "10%",
				},
			},
		},
		{
			name: "User already using rolling_update_strategy field",
			input: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
			existingFields: map[string]interface{}{
				"rolling_update_strategy": []interface{}{},
			},
			expectedLegacy: "RollingUpdateScaleOut",
			expectedRolling: []interface{}{
				map[string]interface{}{
					"type": "RollingUpdateScaleOut",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oi := make(map[string]interface{})
			// Copy existing fields
			for k, v := range tt.existingFields {
				oi[k] = v
			}

			flattenUpdateStrategy(tt.input, oi)

			// Check legacy field
			assert.Equal(t, tt.expectedLegacy, oi["update_strategy"])

			// Check rolling_update_strategy field
			if tt.expectedRolling == nil {
				assert.Nil(t, oi["rolling_update_strategy"])
			} else {
				assert.Equal(t, tt.expectedRolling, oi["rolling_update_strategy"])
			}
		})
	}
}

func TestUpdateStrategyBackwardCompatibility(t *testing.T) {
	t.Run("Legacy field works when rolling_update_strategy is not used", func(t *testing.T) {
		// User using legacy update_strategy
		input := map[string]interface{}{
			"update_strategy": "RollingUpdateScaleIn",
		}

		// Should get the strategy from legacy field
		strategy := toUpdateStrategy(input)
		assert.Equal(t, "RollingUpdateScaleIn", strategy.Type)
		assert.Empty(t, strategy.MaxSurge)
		assert.Empty(t, strategy.MaxUnavailable)
	})

	t.Run("New field takes precedence over legacy field", func(t *testing.T) {
		// User has both fields (migration scenario)
		input := map[string]interface{}{
			"update_strategy": "RollingUpdateScaleIn",
			"rolling_update_strategy": []interface{}{
				map[string]interface{}{
					"type":            "OverrideScaling",
					"max_surge":       "1",
					"max_unavailable": "0",
				},
			},
		}

		// Should use new field
		strategy := toUpdateStrategy(input)
		assert.Equal(t, "OverrideScaling", strategy.Type)
		assert.Equal(t, "1", strategy.MaxSurge)
		assert.Equal(t, "0", strategy.MaxUnavailable)
	})

	t.Run("Flatten sets both fields for backward compatibility", func(t *testing.T) {
		// API returns strategy with max fields
		apiStrategy := &models.V1UpdateStrategy{
			Type:           "OverrideScaling",
			MaxSurge:       "2",
			MaxUnavailable: "1",
		}

		oi := make(map[string]interface{})
		flattenUpdateStrategy(apiStrategy, oi)

		// Both fields should be set
		assert.Equal(t, "OverrideScaling", oi["update_strategy"])
		assert.NotNil(t, oi["rolling_update_strategy"])

		rollingUpdate := oi["rolling_update_strategy"].([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "OverrideScaling", rollingUpdate["type"])
		assert.Equal(t, "2", rollingUpdate["max_surge"])
		assert.Equal(t, "1", rollingUpdate["max_unavailable"])
	})
}
