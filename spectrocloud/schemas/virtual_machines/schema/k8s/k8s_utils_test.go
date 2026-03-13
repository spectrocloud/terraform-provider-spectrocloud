package k8s

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestNamespacedMetadataSchemaIsTemplate(t *testing.T) {
	tests := []struct {
		objectName      string
		generatableName bool
		isTemplate      bool
		expectedFields  map[string]*schema.Schema
	}{
		{
			objectName:      "pod",
			generatableName: true,
			isTemplate:      false,
			expectedFields: map[string]*schema.Schema{
				"namespace": {
					Type:        schema.TypeString,
					Description: "Namespace defines the space within which name of the pod must be unique.",
					Optional:    true,
					ForceNew:    true,
					Default:     "default",
				},
				"generate_name": {
					Type:          schema.TypeString,
					Description:   "Prefix, used by the server, to generate a unique name ONLY IF the `name` field has not been provided. This value will also be combined with a unique suffix. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#idempotency",
					Optional:      true,
					ValidateFunc:  utils.ValidateGenerateName,
					ConflictsWith: []string{"metadata.name"},
				},
				"name": {
					Type:        schema.TypeString,
					Description: "Name of the pod.",
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
		{
			objectName:      "service",
			generatableName: false,
			isTemplate:      true,
			expectedFields: map[string]*schema.Schema{
				"namespace": {
					Type:        schema.TypeString,
					Description: "Namespace defines the space within which name of the service must be unique.",
					Optional:    true,
					ForceNew:    true,
					Default:     nil,
				},
				"generate_name": nil,
				"name": {
					Type:        schema.TypeString,
					Description: "Name of the service.",
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.objectName, func(t *testing.T) {
			_ = namespacedMetadataSchemaIsTemplate(tt.objectName, tt.generatableName, tt.isTemplate)
		})
	}
}

func TestConvertToBasicMetadata(t *testing.T) {
	tests := []struct {
		name         string
		resourceData *schema.ResourceData
		expectedMeta metav1.ObjectMeta
	}{
		{
			name:         "complete metadata",
			resourceData: &schema.ResourceData{
				// Initialize the ResourceData with the necessary values
				// For example, use a mock or set values directly
				// Assuming `schema.ResourceData` has methods like `Set`, `GetOk` etc.
			},
			expectedMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
				},
				Labels: map[string]string{
					"label1": "value1",
				},
				GenerateName:    "gen-name",
				Name:            "name",
				Namespace:       "namespace",
				ResourceVersion: "resource-version",
			},
		},
		{
			name:         "partial metadata",
			resourceData: &schema.ResourceData{
				// Initialize the ResourceData with only some values
			},
			expectedMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
				},
				Labels: map[string]string{
					"label1": "value1",
				},
				Name: "name",
			},
		},
		{
			name:         "empty metadata",
			resourceData: &schema.ResourceData{
				// Initialize the ResourceData with no values
			},
			expectedMeta: metav1.ObjectMeta{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = ConvertToBasicMetadata(tt.resourceData)
		})
	}
}

func TestExpandLabelSelectorRequirement(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []metav1.LabelSelectorRequirement
	}{
		{
			name: "valid requirements",
			input: []interface{}{
				map[string]interface{}{
					"key":      "key1",
					"operator": "In",
					"values":   schema.NewSet(schema.HashString, []interface{}{"value1", "value2"}),
				},
				map[string]interface{}{
					"key":      "key2",
					"operator": "NotIn",
					"values":   schema.NewSet(schema.HashString, []interface{}{"value3"}),
				},
			},
			expected: []metav1.LabelSelectorRequirement{
				{
					Key:      "key1",
					Operator: metav1.LabelSelectorOperator("In"),
					Values:   []string{"value1", "value2"},
				},
				{
					Key:      "key2",
					Operator: metav1.LabelSelectorOperator("NotIn"),
					Values:   []string{"value3"},
				},
			},
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []metav1.LabelSelectorRequirement{},
		},
		{
			name: "nil input",
			input: []interface{}{
				nil,
			},
			expected: []metav1.LabelSelectorRequirement{},
		},
		{
			name: "invalid input",
			input: []interface{}{
				map[string]interface{}{
					"key":      "key1",
					"operator": "InvalidOperator",
					"values":   schema.NewSet(schema.HashString, []interface{}{"value1"}),
				},
			},
			expected: []metav1.LabelSelectorRequirement{
				{
					Key:      "key1",
					Operator: metav1.LabelSelectorOperator("InvalidOperator"),
					Values:   []string{"value1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandLabelSelectorRequirement(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestFlattenLabelSelectorRequirement(t *testing.T) {
	tests := []struct {
		name     string
		input    []metav1.LabelSelectorRequirement
		expected []interface{}
	}{
		{
			name:     "empty input",
			input:    []metav1.LabelSelectorRequirement{},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenLabelSelectorRequirement(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
