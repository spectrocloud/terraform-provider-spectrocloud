package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandMetadata(t *testing.T) {
	tests := []struct {
		name   string
		input  []interface{}
		output *models.V1ObjectMetaInputEntity
	}{
		{
			name:   "empty list",
			input:  []interface{}{},
			output: nil,
		},
		{
			name:   "nil element",
			input:  []interface{}{nil},
			output: nil,
		},
		{
			name: "valid metadata",
			input: []interface{}{
				map[string]interface{}{
					"name": "test",
				},
			},
			output: &models.V1ObjectMetaInputEntity{
				Name: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandMetadata(tt.input)
			if !reflect.DeepEqual(result, tt.output) {
				t.Errorf("expandMetadata() = %v, want %v", result, tt.output)
			}
		})
	}
}

func TestExpandSpec(t *testing.T) {
	conjunction := models.V1SearchFilterConjunctionOperator("and")

	tests := []struct {
		name   string
		input  []interface{}
		output *models.V1TagFilterSpec
	}{
		{
			name:   "empty list",
			input:  []interface{}{},
			output: nil,
		},
		{
			name:   "nil element",
			input:  []interface{}{nil},
			output: nil,
		},
		{
			name: "valid spec",
			input: []interface{}{
				map[string]interface{}{
					"filter_group": []interface{}{
						map[string]interface{}{
							"conjunction": "and",
							"filters": []interface{}{
								map[string]interface{}{
									"key":      "test_key",
									"negation": false,
									"operator": "EQUALS",
									"values":   []interface{}{"test_value"},
								},
							},
						},
					},
				},
			},
			output: &models.V1TagFilterSpec{
				FilterGroup: &models.V1TagFilterGroup{
					Conjunction: &conjunction,
					Filters: []*models.V1TagFilterItem{
						{
							Key:      "test_key",
							Negation: false,
							Operator: models.V1SearchFilterKeyValueOperator("EQUALS"),
							Values:   []string{"test_value"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandSpec(tt.input)
			assert.Equal(t, tt.output, result, "expandSpec() returned unexpected result")
		})
	}
}

func TestFlattenMetadata(t *testing.T) {
	tests := []struct {
		name   string
		input  *models.V1ObjectMeta
		output []interface{}
	}{
		{
			name:   "nil metadata",
			input:  nil,
			output: []interface{}{},
		},
		{
			name: "valid metadata with name",
			input: &models.V1ObjectMeta{
				Name: "test_name",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "test_name",
				},
			},
		},
		{
			name: "metadata with empty name",
			input: &models.V1ObjectMeta{
				Name: "",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "",
				},
			},
		},
		{
			name: "metadata with long name",
			input: &models.V1ObjectMeta{
				Name: "very-long-name-with-many-characters-that-exceeds-normal-length",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "very-long-name-with-many-characters-that-exceeds-normal-length",
				},
			},
		},
		{
			name: "metadata with special characters in name",
			input: &models.V1ObjectMeta{
				Name: "test-name_with.special@chars#123",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "test-name_with.special@chars#123",
				},
			},
		},
		{
			name: "metadata with numeric name",
			input: &models.V1ObjectMeta{
				Name: "12345",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "12345",
				},
			},
		},
		{
			name: "metadata with name containing spaces",
			input: &models.V1ObjectMeta{
				Name: "test name with spaces",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "test name with spaces",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenMetadata(tt.input)

			require.Equal(t, len(tt.output), len(result), "Result length should match expected")

			if len(tt.output) > 0 {
				expectedMap := tt.output[0].(map[string]interface{})
				resultMap := result[0].(map[string]interface{})
				assert.Equal(t, expectedMap["name"], resultMap["name"], "Name should match")
			} else {
				assert.Empty(t, result, "Result should be empty for nil metadata")
			}
		})
	}
}

func TestFlattenSpec(t *testing.T) {
	conjunction := models.V1SearchFilterConjunctionOperator("and")
	tests := []struct {
		name   string
		input  *models.V1TagFilterSpec
		verify func(t *testing.T, result []interface{})
	}{
		{
			name:  "nil spec",
			input: nil,
			verify: func(t *testing.T, result []interface{}) {
				assert.Equal(t, []interface{}{}, result)
			},
		},
		{
			name: "valid spec",
			input: &models.V1TagFilterSpec{
				FilterGroup: &models.V1TagFilterGroup{
					Conjunction: &conjunction,
					Filters: []*models.V1TagFilterItem{
						{
							Key:      "test_key",
							Negation: false,
							Operator: models.V1SearchFilterKeyValueOperator("EQUALS"),
							Values:   []string{"test_value"},
						},
					},
				},
			},
			verify: func(t *testing.T, result []interface{}) {
				// Verify structure
				assert.Len(t, result, 1)
				resultMap := result[0].(map[string]interface{})

				// Verify filter_group
				filterGroupList := resultMap["filter_group"].([]interface{})
				assert.Len(t, filterGroupList, 1)
				filterGroupMap := filterGroupList[0].(map[string]interface{})
				assert.Equal(t, "and", filterGroupMap["conjunction"])

				// Verify filters is []interface{} (works with both TypeSet and TypeList)
				filtersList, ok := filterGroupMap["filters"].([]interface{})
				assert.True(t, ok, "filters should be []interface{}")
				assert.Len(t, filtersList, 1)

				// Verify filter content
				filterMap := filtersList[0].(map[string]interface{})
				assert.Equal(t, "test_key", filterMap["key"])
				assert.Equal(t, false, filterMap["negation"])
				assert.Equal(t, "EQUALS", filterMap["operator"])
				assert.Equal(t, []string{"test_value"}, filterMap["values"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenSpec(tt.input)
			tt.verify(t, result)
		})
	}
}

func TestFlattenFilterGroup(t *testing.T) {
	tests := []struct {
		name   string
		input  *models.V1TagFilterGroup
		output []interface{}
	}{
		{
			name:   "nil filter group",
			input:  nil,
			output: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenFilterGroup(tt.input)
			assert.Equal(t, tt.output, result)
		})
	}
}

func TestExpandFilterGroup(t *testing.T) {
	tests := []struct {
		name   string
		input  []interface{}
		output *models.V1TagFilterGroup
	}{
		{
			name:   "empty input",
			input:  []interface{}{},
			output: nil,
		},
		{
			name:   "nil input",
			input:  []interface{}{nil},
			output: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandFilterGroup(tt.input)
			assert.Equal(t, tt.output, result)
		})
	}
}

func TestFlattenFilters(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*models.V1TagFilterItem
		expected []interface{}
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name:     "Empty slice",
			input:    []*models.V1TagFilterItem{},
			expected: []interface{}{},
		},
		{
			name: "Single filter item",
			input: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("EQUALS"),
					Values:   []string{"production"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "EQUALS",
					"values":   []string{"production"},
				},
			},
		},
		{
			name: "Multiple filter items",
			input: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("EQUALS"),
					Values:   []string{"production"},
				},
				{
					Key:      "app",
					Negation: true,
					Operator: models.V1SearchFilterKeyValueOperator("NOT_EQUALS"),
					Values:   []string{"test"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "EQUALS",
					"values":   []string{"production"},
				},
				map[string]interface{}{
					"key":      "app",
					"negation": true,
					"operator": "NOT_EQUALS",
					"values":   []string{"test"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := flattenFilters(tc.input)
			require.ElementsMatch(t, tc.expected, actual)
		})
	}
}

func TestExpandFilters(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*models.V1TagFilterItem
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: []*models.V1TagFilterItem{},
		},
		{
			name: "single filter with values",
			input: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"production"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"production"},
				},
			},
		},
		{
			name: "single filter with nil values",
			input: []interface{}{
				map[string]interface{}{
					"key":      "app",
					"negation": true,
					"operator": "eq",
					"values":   nil,
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "app",
					Negation: true,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   nil,
				},
			},
		},
		{
			name: "single filter with empty values",
			input: []interface{}{
				map[string]interface{}{
					"key":      "tag",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "tag",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   nil, // Empty slice results in nil when loop doesn't execute
				},
			},
		},
		{
			name: "single filter with multiple values",
			input: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"production", "staging", "development"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"production", "staging", "development"},
				},
			},
		},
		{
			name: "multiple filters",
			input: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"production"},
				},
				map[string]interface{}{
					"key":      "app",
					"negation": true,
					"operator": "eq",
					"values":   []interface{}{"test"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"production"},
				},
				{
					Key:      "app",
					Negation: true,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"test"},
				},
			},
		},
		{
			name: "multiple filters with different operators",
			input: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"production"},
				},
				map[string]interface{}{
					"key":      "version",
					"negation": false,
					"operator": "EQUALS",
					"values":   []interface{}{"1.0.0"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"production"},
				},
				{
					Key:      "version",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("EQUALS"),
					Values:   []string{"1.0.0"},
				},
			},
		},
		{
			name: "filter with negation true",
			input: []interface{}{
				map[string]interface{}{
					"key":      "status",
					"negation": true,
					"operator": "eq",
					"values":   []interface{}{"deleted"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "status",
					Negation: true,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"deleted"},
				},
			},
		},
		{
			name: "filter with negation false",
			input: []interface{}{
				map[string]interface{}{
					"key":      "status",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"active"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "status",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"active"},
				},
			},
		},
		{
			name: "filter with single value",
			input: []interface{}{
				map[string]interface{}{
					"key":      "team",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"backend"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "team",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"backend"},
				},
			},
		},
		{
			name: "three filters with various configurations",
			input: []interface{}{
				map[string]interface{}{
					"key":      "env",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"production"},
				},
				map[string]interface{}{
					"key":      "app",
					"negation": true,
					"operator": "eq",
					"values":   nil,
				},
				map[string]interface{}{
					"key":      "region",
					"negation": false,
					"operator": "eq",
					"values":   []interface{}{"us-east-1", "us-west-2"},
				},
			},
			expected: []*models.V1TagFilterItem{
				{
					Key:      "env",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"production"},
				},
				{
					Key:      "app",
					Negation: true,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   nil,
				},
				{
					Key:      "region",
					Negation: false,
					Operator: models.V1SearchFilterKeyValueOperator("eq"),
					Values:   []string{"us-east-1", "us-west-2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandFilters(tt.input)

			require.Equal(t, len(tt.expected), len(result), "Result length should match expected")

			for i, expected := range tt.expected {
				assert.Equal(t, expected.Key, result[i].Key, "Key should match")
				assert.Equal(t, expected.Negation, result[i].Negation, "Negation should match")
				assert.Equal(t, expected.Operator, result[i].Operator, "Operator should match")
				assert.Equal(t, expected.Values, result[i].Values, "Values should match")
			}
		})
	}
}
