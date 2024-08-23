package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
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
		output []interface{}
	}{
		{
			name:   "nil spec",
			input:  nil,
			output: []interface{}{},
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
			output: []interface{}{
				map[string]interface{}{
					"filter_group": []interface{}{
						map[string]interface{}{
							"conjunction": "and",
							"filters": []interface{}{
								map[string]interface{}{
									"key":      "test_key",
									"negation": false,
									"operator": "EQUALS",
									"values":   []string{"test_value"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenSpec(tt.input)
			assert.Equal(t, tt.output, result)
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
