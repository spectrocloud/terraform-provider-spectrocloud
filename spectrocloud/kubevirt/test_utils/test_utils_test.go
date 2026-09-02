package test_utils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestNullifySchemaSetFunction(t *testing.T) {
	ss := schema.NewSet(schema.HashString, []interface{}{"a", "b"})
	a := assert.New(t)
	a.NotNil(ss.F)

	NullifySchemaSetFunction(ss)

	a.Nil(ss.F)
	// Set contents/behavior are unaffected by nullifying F.
	a.Equal(2, ss.Len())
}

func TestGetPVCRequirements(t *testing.T) {
	tests := []struct {
		name       string
		dataVolume interface{}
		expected   interface{}
	}{
		{
			name: "Valid input",
			dataVolume: map[string]interface{}{
				"spec": []interface{}{
					map[string]interface{}{
						"pvc": []interface{}{
							map[string]interface{}{
								"resources": []interface{}{
									map[string]interface{}{
										"requests": map[string]interface{}{
											"storage": "10Gi",
										},
										"limits": map[string]interface{}{
											"storage": "20Gi",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]interface{}{
				"requests": map[string]interface{}{
					"storage": "10Gi",
				},
				"limits": map[string]interface{}{
					"storage": "20Gi",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPVCRequirements(tt.dataVolume)
			assert.Equal(t, tt.expected, result)
		})
	}
}
