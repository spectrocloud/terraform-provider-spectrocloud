package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func expandMetadata(list []interface{}) *models.V1ObjectMetaInputEntity {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	m := list[0].(map[string]interface{})

	return &models.V1ObjectMetaInputEntity{
		Name: m["name"].(string),
	}
}

func expandSpec(list []interface{}) *models.V1TagFilterSpec {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	m := list[0].(map[string]interface{})
	filterGroup := m["filter_group"].([]interface{})

	return &models.V1TagFilterSpec{
		FilterGroup: expandFilterGroup(filterGroup),
	}
}

func expandFilterGroup(list []interface{}) *models.V1TagFilterGroup {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	m := list[0].(map[string]interface{})

	// Handle both TypeSet and TypeList for backward compatibility
	var filtersList []interface{}
	if filtersRaw, ok := m["filters"]; ok && filtersRaw != nil {
		if filtersSet, ok := filtersRaw.(*schema.Set); ok {
			filtersList = filtersSet.List()
		} else if filtersListRaw, ok := filtersRaw.([]interface{}); ok {
			// Fallback for backward compatibility during migration
			filtersList = filtersListRaw
		}
	}

	// filters := m["filters"].([]interface{})

	conjunction := models.V1SearchFilterConjunctionOperator(m["conjunction"].(string))

	return &models.V1TagFilterGroup{
		Conjunction: &conjunction,
		Filters:     expandFilters(filtersList),
	}
}

func expandFilters(list []interface{}) []*models.V1TagFilterItem {
	filters := make([]*models.V1TagFilterItem, len(list))

	for i, item := range list {
		m := item.(map[string]interface{})
		var values []string
		if m["values"] != nil {
			interfaceValues := m["values"].([]interface{})
			for _, v := range interfaceValues {
				values = append(values, v.(string))
			}
		}

		filters[i] = &models.V1TagFilterItem{
			Key:      m["key"].(string),
			Negation: m["negation"].(bool),
			Operator: models.V1SearchFilterKeyValueOperator(m["operator"].(string)),
			Values:   values,
		}
	}

	return filters
}

func flattenMetadata(metadata *models.V1ObjectMeta) []interface{} {
	if metadata == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"name": metadata.Name,
	}

	return []interface{}{m}
}

func flattenSpec(spec *models.V1TagFilterSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"filter_group": flattenFilterGroup(spec.FilterGroup),
	}

	return []interface{}{m}
}

func flattenFilters(filters []*models.V1TagFilterItem) []interface{} {
	if filters == nil {
		return []interface{}{}
	}

	fs := make([]interface{}, len(filters))

	for i, filter := range filters {
		m := map[string]interface{}{
			"key":      filter.Key,
			"negation": filter.Negation,
			"operator": string(filter.Operator),
			"values":   filter.Values,
		}
		fs[i] = m
	}

	return fs
}

func flattenFilterGroup(filterGroup *models.V1TagFilterGroup) []interface{} {
	if filterGroup == nil {
		return []interface{}{}
	}

	// Return filters as []interface{} - Terraform will handle conversion
	// to *schema.Set automatically if the schema is TypeSet
	var filtersList []interface{}
	if filterGroup.Filters != nil && len(filterGroup.Filters) > 0 {
		filtersList = make([]interface{}, len(filterGroup.Filters))
		for i, filter := range filterGroup.Filters {
			filtersList[i] = map[string]interface{}{
				"key":      filter.Key,
				"negation": filter.Negation,
				"operator": string(filter.Operator),
				"values":   filter.Values,
			}
		}
	}

	conjunction := ""
	if filterGroup.Conjunction != nil {
		conjunction = string(*filterGroup.Conjunction)
	}

	m := map[string]interface{}{
		"conjunction": conjunction,
		"filters":     filtersList, // Return []interface{} instead of *schema.Set
	}

	return []interface{}{m}
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
