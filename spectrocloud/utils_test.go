package spectrocloud

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringContains(t *testing.T) {
	ss := []string{"fizz_1", "bazz", "random", "nfizz_1", "fizz_2"}

	contains := stringContains(ss, "random")
	assert.Equal(t, true, contains, "Should be true.")

	assert.Equal(t, false, stringContains(ss, "doesnt"), "Should be false.")
}

func TestFilter(t *testing.T) {
	ss := []string{"fizz_1", "bazz", "random", "nfizz_1", "fizz_2"}

	mytest := func(s string) bool { return !strings.HasPrefix(s, "fizz_") && len(s) <= 7 }
	s3 := filter(ss, mytest)

	assert.Equal(t, 3, len(s3), "The two len should be the same.")
}

func TestIsMapSubset(t *testing.T) {
	a := map[string]string{"a": "b", "c": "d", "e": "f"}
	b := map[string]string{"a": "b", "e": "f"}
	c := map[string]string{"a": "b", "e": "g"}

	assert.Equal(t, true, IsMapSubset(a, b))
	assert.Equal(t, false, IsMapSubset(a, c))
	assert.Equal(t, false, IsMapSubset(b, a)) // a bigger than b
}

func TestSafeInt32(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    int32
		description string
	}{
		{
			name:        "Normal value within int32 range",
			input:       100,
			expected:    100,
			description: "Should convert normal int value to int32",
		},
		{
			name:        "Zero value",
			input:       0,
			expected:    0,
			description: "Should handle zero value",
		},
		{
			name:        "Negative value within range",
			input:       -100,
			expected:    -100,
			description: "Should handle negative values within int32 range",
		},
		{
			name:        "MaxInt32 boundary",
			input:       int(math.MaxInt32),
			expected:    math.MaxInt32,
			description: "Should handle MaxInt32 boundary value",
		},
		{
			name:        "MinInt32 boundary",
			input:       int(math.MinInt32),
			expected:    math.MinInt32,
			description: "Should handle MinInt32 boundary value",
		},
		{
			name:        "Value exceeding MaxInt32",
			input:       int(math.MaxInt32) + 1,
			expected:    math.MaxInt32,
			description: "Should clamp to MaxInt32 when value exceeds limit",
		},
		{
			name:        "Value below MinInt32",
			input:       int(math.MinInt32) - 1,
			expected:    math.MinInt32,
			description: "Should clamp to MinInt32 when value is below limit",
		},
		{
			name:        "Very large positive value",
			input:       math.MaxInt,
			expected:    math.MaxInt32,
			description: "Should clamp very large positive value to MaxInt32",
		},
		{
			name:        "Very large negative value",
			input:       math.MinInt,
			expected:    math.MinInt32,
			description: "Should clamp very large negative value to MinInt32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeInt32(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestSafeInt64(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expected    int64
		description string
	}{
		{
			name:        "Normal value",
			input:       100,
			expected:    100,
			description: "Should convert normal int value to int64",
		},
		{
			name:        "Zero value",
			input:       0,
			expected:    0,
			description: "Should handle zero value",
		},
		{
			name:        "Negative value",
			input:       -100,
			expected:    -100,
			description: "Should handle negative values",
		},
		{
			name:        "MaxInt32 value",
			input:       int(math.MaxInt32),
			expected:    int64(math.MaxInt32),
			description: "Should convert MaxInt32 to int64",
		},
		{
			name:        "MinInt32 value",
			input:       int(math.MinInt32),
			expected:    int64(math.MinInt32),
			description: "Should convert MinInt32 to int64",
		},
		{
			name:        "MaxInt value",
			input:       math.MaxInt,
			expected:    int64(math.MaxInt),
			description: "Should convert MaxInt to int64",
		},
		{
			name:        "MinInt value",
			input:       math.MinInt,
			expected:    int64(math.MinInt),
			description: "Should convert MinInt to int64",
		},
		{
			name:        "Value exceeding MaxInt32",
			input:       int(math.MaxInt32) + 1,
			expected:    int64(math.MaxInt32) + 1,
			description: "Should convert values exceeding MaxInt32",
		},
		{
			name:        "Value below MinInt32",
			input:       int(math.MinInt32) - 1,
			expected:    int64(math.MinInt32) - 1,
			description: "Should convert values below MinInt32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeInt64(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestExpandStringList(t *testing.T) {
	tests := []struct {
		name        string
		input       []interface{}
		expected    []string
		description string
	}{
		{
			name:        "Empty slice",
			input:       []interface{}{},
			expected:    []string{},
			description: "Should return empty slice for empty input",
		},
		{
			name:        "Single string",
			input:       []interface{}{"test"},
			expected:    []string{"test"},
			description: "Should convert single string correctly",
		},
		{
			name:        "Multiple strings",
			input:       []interface{}{"test1", "test2", "test3"},
			expected:    []string{"test1", "test2", "test3"},
			description: "Should convert multiple strings correctly",
		},
		{
			name:        "Slice with nil values",
			input:       []interface{}{"test1", nil, "test2", nil, "test3"},
			expected:    []string{"test1", "test2", "test3"},
			description: "Should skip nil values and return only strings",
		},
		{
			name:        "Slice with only nil values",
			input:       []interface{}{nil, nil, nil},
			expected:    []string{},
			description: "Should return empty slice when all values are nil",
		},
		{
			name:        "Slice with empty strings",
			input:       []interface{}{"", "test", ""},
			expected:    []string{"", "test", ""},
			description: "Should preserve empty strings",
		},
		{
			name:        "Slice with long strings",
			input:       []interface{}{"very-long-string-with-many-characters", "another-long-string"},
			expected:    []string{"very-long-string-with-many-characters", "another-long-string"},
			description: "Should handle long strings",
		},
		{
			name:        "Slice with strings containing spaces",
			input:       []interface{}{"test string", "another test", "final test"},
			expected:    []string{"test string", "another test", "final test"},
			description: "Should handle strings with spaces",
		},
		{
			name:        "Alternating nil and strings",
			input:       []interface{}{"test1", nil, "test2", nil, "test3", nil},
			expected:    []string{"test1", "test2", "test3"},
			description: "Should handle alternating nil and string values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandStringList(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestExpandStringMap(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		expected    map[string]string
		description string
	}{
		{
			name:        "Empty map",
			input:       map[string]interface{}{},
			expected:    map[string]string{},
			description: "Should return empty map for empty input",
		},
		{
			name:        "Single key-value pair",
			input:       map[string]interface{}{"key1": "value1"},
			expected:    map[string]string{"key1": "value1"},
			description: "Should convert single key-value pair correctly",
		},
		{
			name:        "Multiple key-value pairs",
			input:       map[string]interface{}{"key1": "value1", "key2": "value2", "key3": "value3"},
			expected:    map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
			description: "Should convert multiple key-value pairs correctly",
		},
		{
			name:        "Map with special characters in keys",
			input:       map[string]interface{}{"key-1": "value1", "key_2": "value2", "key@3": "value3"},
			expected:    map[string]string{"key-1": "value1", "key_2": "value2", "key@3": "value3"},
			description: "Should handle special characters in keys",
		},
		{
			name:        "Map with special characters in values",
			input:       map[string]interface{}{"key1": "value-1", "key2": "value_2", "key3": "value@3"},
			expected:    map[string]string{"key1": "value-1", "key2": "value_2", "key3": "value@3"},
			description: "Should handle special characters in values",
		},
		{
			name:        "Map with long strings",
			input:       map[string]interface{}{"key1": "very-long-string-with-many-characters", "key2": "another-long-string"},
			expected:    map[string]string{"key1": "very-long-string-with-many-characters", "key2": "another-long-string"},
			description: "Should handle long string values",
		},
		{
			name:        "Map with numeric string values",
			input:       map[string]interface{}{"key1": "123", "key2": "456", "key3": "789"},
			expected:    map[string]string{"key1": "123", "key2": "456", "key3": "789"},
			description: "Should handle numeric string values",
		},
		{
			name:        "Map with boolean string values",
			input:       map[string]interface{}{"key1": "true", "key2": "false"},
			expected:    map[string]string{"key1": "true", "key2": "false"},
			description: "Should handle boolean string values",
		},
		{
			name:        "Map with single character keys and values",
			input:       map[string]interface{}{"a": "b", "c": "d", "e": "f"},
			expected:    map[string]string{"a": "b", "c": "d", "e": "f"},
			description: "Should handle single character keys and values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandStringMap(tt.input)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestInt16WithDefault(t *testing.T) {
	tests := []struct {
		name        string
		input       *int16
		defaultVal  int16
		expected    int16
		description string
	}{
		{
			name:        "Nil pointer with default value",
			input:       nil,
			defaultVal:  100,
			expected:    100,
			description: "Should return default value when pointer is nil",
		},
		{
			name:        "Nil pointer with zero default",
			input:       nil,
			defaultVal:  0,
			expected:    0,
			description: "Should return zero default when pointer is nil",
		},
		{
			name:        "Nil pointer with negative default",
			input:       nil,
			defaultVal:  -100,
			expected:    -100,
			description: "Should return negative default when pointer is nil",
		},
		{
			name:        "Valid pointer with value",
			input:       int16Ptr(50),
			defaultVal:  100,
			expected:    50,
			description: "Should return pointer value when not nil",
		},
		{
			name:        "Valid pointer with zero value",
			input:       int16Ptr(0),
			defaultVal:  100,
			expected:    0,
			description: "Should return zero value from pointer, not default",
		},
		{
			name:        "Valid pointer with negative value",
			input:       int16Ptr(-50),
			defaultVal:  100,
			expected:    -50,
			description: "Should return negative value from pointer",
		},
		{
			name:        "Valid pointer with MaxInt16",
			input:       int16Ptr(math.MaxInt16),
			defaultVal:  100,
			expected:    math.MaxInt16,
			description: "Should return MaxInt16 from pointer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int16WithDefault(tt.input, tt.defaultVal)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// Helper function for creating int16 pointers
func int16Ptr(i int16) *int16 {
	return &i
}

func TestStringWithDefaultValue(t *testing.T) {
	tests := []struct {
		name        string
		input       *string
		defaultVal  string
		expected    string
		description string
	}{
		{
			name:        "Nil pointer with default value",
			input:       nil,
			defaultVal:  "default",
			expected:    "default",
			description: "Should return default value when pointer is nil",
		},
		{
			name:        "Nil pointer with empty default",
			input:       nil,
			defaultVal:  "",
			expected:    "",
			description: "Should return empty default when pointer is nil",
		},
		{
			name:        "Nil pointer with long default",
			input:       nil,
			defaultVal:  "very-long-default-string-with-many-characters",
			expected:    "very-long-default-string-with-many-characters",
			description: "Should return long default when pointer is nil",
		},
		{
			name:        "Valid pointer with value",
			input:       stringPtr("actual"),
			defaultVal:  "default",
			expected:    "actual",
			description: "Should return pointer value when not nil",
		},
		{
			name:        "Valid pointer with empty string",
			input:       stringPtr(""),
			defaultVal:  "default",
			expected:    "",
			description: "Should return empty string from pointer, not default",
		},
		{
			name:        "Valid pointer value equals default",
			input:       stringPtr("default"),
			defaultVal:  "default",
			expected:    "default",
			description: "Should return pointer value even when it equals default",
		},
		{
			name:        "Nil pointer with spaces in default",
			input:       nil,
			defaultVal:  "default with spaces",
			expected:    "default with spaces",
			description: "Should return default with spaces when pointer is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringWithDefaultValue(tt.input, tt.defaultVal)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
