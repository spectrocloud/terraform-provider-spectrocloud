package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

func TestValidateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		key      string
		expected []error
	}{
		{
			name: "Valid annotations",
			value: map[string]interface{}{
				"valid.annotation/key": "value",
				"another.valid/key":    "value",
			},
			key:      "annotations",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, es := ValidateAnnotations(tt.value, tt.key)
			assert.Equal(t, tt.expected, es)
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		key      string
		expected []error
	}{
		{
			name:     "Valid name",
			value:    "valid-name",
			key:      "name",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, es := ValidateName(tt.value, tt.key)
			assert.Equal(t, tt.expected, es)
		})
	}
}

func TestValidateGenerateName(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		key      string
		expected []error
	}{
		{
			name:     "Valid generate name",
			value:    "valid-name",
			key:      "generate_name",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, es := ValidateGenerateName(tt.value, tt.key)
			assert.Equal(t, tt.expected, es)
		})
	}
}

func TestValidateLabels(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		key      string
		expected []error
	}{
		{
			name: "Valid labels",
			value: map[string]interface{}{
				"valid.label/key": "valid-value",
			},
			key:      "labels",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, es := ValidateLabels(tt.value, tt.key)
			assert.Equal(t, tt.expected, es)
		})
	}
}

func TestValidateTypeStringNullableInt(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		key      string
		expected []error
	}{
		{
			name:     "Valid integer",
			value:    "123",
			key:      "nullable_int",
			expected: nil,
		},
		{
			name:     "Empty string",
			value:    "",
			key:      "nullable_int",
			expected: nil,
		},
		{
			name:  "Invalid string",
			value: "abc",
			key:   "nullable_int",
			expected: []error{
				fmt.Errorf("nullable_int: cannot parse 'abc' as int: strconv.ParseInt: parsing \"abc\": invalid syntax"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, es := ValidateTypeStringNullableInt(tt.value, tt.key)
			assert.Equal(t, tt.expected, es)
		})
	}
}

func TestStringIsIntInRange(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		min      int
		max      int
		expected diag.Diagnostics
	}{
		{
			name:     "Valid integer within range",
			value:    "5",
			min:      1,
			max:      10,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagFunc := StringIsIntInRange(tt.min, tt.max)
			result := diagFunc(tt.value, cty.Path{})
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIdParts(t *testing.T) {
	tests := []struct {
		id       string
		expected [4]string
		hasError bool
	}{
		{"scope/uid/ns/name", [4]string{"scope", "uid", "ns", "name"}, false},
		{"invalid/id/format", [4]string{"", "", "", ""}, true},
	}

	for _, test := range tests {
		scope, uid, ns, name, err := IdParts(test.id)
		if test.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, [4]string{scope, uid, ns, name})
		}
	}
}

func TestIdPartsDV(t *testing.T) {
	tests := []struct {
		id       string
		expected [5]string
		hasError bool
	}{
		{"scope/uid/ns/name/dv", [5]string{"scope", "uid", "ns", "name", "dv"}, false},
		{"invalid/id/format", [5]string{"", "", "", "", ""}, true},
	}

	for _, test := range tests {
		scope, uid, ns, name, dv, err := IdPartsDV(test.id)
		if test.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, [5]string{scope, uid, ns, name, dv})
		}
	}
}

func TestFlattenStringMap(t *testing.T) {
	m := map[string]string{"key": "value"}
	result := FlattenStringMap(m)
	assert.Equal(t, map[string]interface{}{"key": "value"}, result)
}

func TestExpandStringMap(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	result := ExpandStringMap(m)
	assert.Equal(t, map[string]string{"key": "value"}, result)
}

func TestExpandBase64MapToByteMap(t *testing.T) {
	m := map[string]interface{}{"key": base64.StdEncoding.EncodeToString([]byte("value"))}
	result := ExpandBase64MapToByteMap(m)
	assert.Equal(t, map[string][]byte{"key": []byte("value")}, result)
}

func TestExpandStringMapToByteMap(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	result := ExpandStringMapToByteMap(m)
	assert.Equal(t, map[string][]byte{"key": []byte("value")}, result)
}

func TestExpandStringSlice(t *testing.T) {
	s := []interface{}{"a", "b", "c"}
	result := ExpandStringSlice(s)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestFlattenByteMapToBase64Map(t *testing.T) {
	m := map[string][]byte{"key": []byte("value")}
	result := FlattenByteMapToBase64Map(m)
	assert.Equal(t, map[string]string{"key": base64.StdEncoding.EncodeToString([]byte("value"))}, result)
}

func TestFlattenByteMapToStringMap(t *testing.T) {
	m := map[string][]byte{"key": []byte("value")}
	result := FlattenByteMapToStringMap(m)
	assert.Equal(t, map[string]string{"key": "value"}, result)
}

func TestPtrToString(t *testing.T) {
	s := "value"
	ptr := PtrToString(s)
	assert.Equal(t, &s, ptr)
}

func TestPtrToBool(t *testing.T) {
	b := true
	ptr := PtrToBool(b)
	assert.Equal(t, &b, ptr)
}

func TestPtrToInt32(t *testing.T) {
	i := int32(10)
	ptr := PtrToInt32(i)
	assert.Equal(t, &i, ptr)
}

func TestPtrToInt64(t *testing.T) {
	i := int64(10)
	ptr := PtrToInt64(i)
	assert.Equal(t, &i, ptr)
}

func TestSliceOfString(t *testing.T) {
	s := []interface{}{"a", "b", "c"}
	result := SliceOfString(s)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestBase64EncodeStringMap(t *testing.T) {
	m := map[string]interface{}{"key": "value"}
	result := Base64EncodeStringMap(m)
	assert.Equal(t, map[string]interface{}{"key": base64.StdEncoding.EncodeToString([]byte("value"))}, result)
}

func TestNewInt64Set(t *testing.T) {
	in := []int64{3, 1, 2}
	set := NewInt64Set(schema.HashInt, in)
	assert.Equal(t, 3, len(set.List()))
}

func TestSchemaSetToStringArray(t *testing.T) {
	set := schema.NewSet(schema.HashString, []interface{}{"a", "b", "c"})
	result := SchemaSetToStringArray(set)
	assert.Equal(t, 3, len(result))
}

func TestExpandMapToResourceList(t *testing.T) {
	m := map[string]interface{}{"cpu": "100m", "memory": "200Mi"}
	rl, err := ExpandMapToResourceList(m)
	assert.NoError(t, err)
	assert.Equal(t, api.ResourceList{
		"cpu":    resource.MustParse("100m"),
		"memory": resource.MustParse("200Mi"),
	}, *rl)
}

func TestFlattenResourceList(t *testing.T) {
	rl := api.ResourceList{
		"cpu":    resource.MustParse("100m"),
		"memory": resource.MustParse("200Mi"),
	}
	m := FlattenResourceList(rl)
	assert.Equal(t, map[string]string{
		"cpu":    "100m",
		"memory": "200Mi",
	}, m)
}
