package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			name: "valid metadata",
			input: &models.V1ObjectMeta{
				Name: "test_name",
			},
			output: []interface{}{
				map[string]interface{}{
					"name": "test_name",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenMetadata(tt.input)
			assert.Equal(t, tt.output, result)
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

				// Verify filters is a *schema.Set
				filtersSet, ok := filterGroupMap["filters"].(*schema.Set)
				assert.True(t, ok, "filters should be *schema.Set")
				assert.Equal(t, 1, filtersSet.Len())

				// Verify filter content
				filterList := filtersSet.List()
				assert.Len(t, filterList, 1)
				filterMap := filterList[0].(map[string]interface{})
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
